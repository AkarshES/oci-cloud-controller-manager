package client

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
)

type mockServiceError struct {
	StatusCode   int
	Code         string
	Message      string
	OpcRequestID string
}

func (m mockServiceError) GetHTTPStatusCode() int {
	return m.StatusCode
}

func (m mockServiceError) GetMessage() string {
	return m.Message
}

func (m mockServiceError) GetCode() string {
	return m.Code
}

func (m mockServiceError) GetOpcRequestID() string {
	return m.OpcRequestID
}
func (m mockServiceError) Error() string {
	return m.Message
}

func TestIsRetryableServiceError(t *testing.T) {
	testCases := map[string]struct {
		error    common.ServiceError
		expected bool
	}{
		"HTTP400RelatedResourceNotAuthorizedOrNotFound": {
			error: mockServiceError{
				StatusCode: http.StatusBadRequest,
				Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
			},
			expected: true,
		},
		"HTTP401NotAuthenticated": {
			error: mockServiceError{
				StatusCode: http.StatusUnauthorized,
				Code:       HTTP401NotAuthenticatedCode,
			},
			expected: true,
		},
		"HTTP404NotAuthorizedOrNotFound": {
			error: mockServiceError{
				StatusCode: http.StatusNotFound,
				Code:       HTTP404NotAuthorizedOrNotFoundCode,
			},
			expected: true,
		},
		"HTTP409IncorrectState": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP409IncorrectStateCode,
			},
			expected: true,
		},
		"HTTP409NotAuthorizedOrResourceAlreadyExists": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP409NotAuthorizedOrResourceAlreadyExistsCode,
			},
			expected: true,
		},
		"HTTP429TooManyRequests": {
			error: mockServiceError{
				StatusCode: http.StatusTooManyRequests,
				Code:       HTTP429TooManyRequestsCode,
			},
			expected: true,
		},
		"HTTP500InternalServerError": {
			error: mockServiceError{
				StatusCode: http.StatusInternalServerError,
				Code:       HTTP500InternalServerErrorCode,
			},
			expected: true,
		},
		"NonRetryable": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP500InternalServerErrorCode,
			},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := isRetryableServiceError(tc.error)
			if result != tc.expected {
				t.Errorf("isRetryableServiceError(%v) = %v ; wanted %v", tc.error, result, tc.expected)
			}
		})
	}

}

func TestIsSystemTagNotFoundOrNotAuthorisedError(t *testing.T) {
	systemTagError := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "The following tag namespaces / keys are not authorized or not found: 'orcl-containerengine'",
	}
	systemTagError2 := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "The following tag namespaces / keys are not authorized or not found: TagDefinition cluster_foobar in TagNamespace orcl-containerengine does not exists.\\n",
	}
	userDefinedTagError1 := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "The following tag namespaces / keys are not authorized or not found: 'foobar-namespace'",
	}
	userDefinedTagError2 := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "TagNamespace orcl-foobar does not exists.\\nTagNamespace orcl-foobar-name does not exists.\\n",
	}
	tests := map[string]struct {
		se               mockServiceError
		wrappedError     error
		expectIsTagError bool
	}{
		"base case": {
			wrappedError:     errors.WithMessage(systemTagError, "taggin failure"),
			expectIsTagError: true,
		},
		"three layer wrapping - resource tracking system tag error": {
			wrappedError:     errors.Wrap(errors.Wrap(errors.WithMessage(systemTagError, "taggin failure"), "first layer"), "second layer"),
			expectIsTagError: true,
		},
		"wrapping with stack trace - resource tracking system tag error": {
			wrappedError:     errors.WithStack(errors.Wrap(errors.WithMessage(systemTagError2, "taggin failure"), "first layer")),
			expectIsTagError: true,
		},
		"three layer wrapping - user defined tag error": {
			wrappedError:     errors.Wrap(errors.Wrap(errors.WithMessage(userDefinedTagError1, "taggin failure"), "first layer"), "second layer"),
			expectIsTagError: false,
		},
		"wrapping with stack trace - user defined tag error": {
			wrappedError:     errors.WithStack(errors.Wrap(errors.WithMessage(userDefinedTagError2, "taggin failure"), "first layer")),
			expectIsTagError: false,
		},
		"not a service error": {
			wrappedError:     errors.Wrap(fmt.Errorf("not a service error"), "precheck error"),
			expectIsTagError: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualResult := IsSystemTagNotFoundOrNotAuthorisedError(zap.S(), test.wrappedError)
			if actualResult != test.expectIsTagError {
				t.Errorf("expected %t but got %t", actualResult, test.expectIsTagError)
			}
		})
	}
}

// testing resource for mocking responses
type mockedResponse struct {
	RawResponse *http.Response
}

// HTTPResponse implements the OCIResponse interface
func (response mockedResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

func mockedOCIOperationResponse(statusCode int, attemptNumber uint) common.OCIOperationResponse {
	httpResponse := http.Response{
		Header:     http.Header{},
		StatusCode: statusCode,
	}
	response := mockedResponse{
		RawResponse: &httpResponse,
	}
	now := time.Now().Round(0)
	return common.NewOCIOperationResponseExtended(response, nil, attemptNumber, &now, 1.0, now)
}

func mockOCIOperationResponseWithError(status int, statusCode string) common.OCIOperationResponse {
	return mockOCIOperationResponseWithErrorFull(status, statusCode, (*time.Time)(nil), 1.0)
}

func mockOCIOperationResponseWithErrorFull(status int, statusCode string, endOfWindowTime *time.Time, backoffScalingFactor float64) common.OCIOperationResponse {
	httpResponse := http.Response{
		Header:     http.Header{},
		StatusCode: status,
	}
	response := mockedResponse{
		RawResponse: &httpResponse,
	}

	err := mockServiceError{
		StatusCode: status,
		Code:       statusCode,
	}

	return common.NewOCIOperationResponseExtended(response, err, uint(1), endOfWindowTime, backoffScalingFactor, time.Now().Round(0))
}

func TestNewRetryPolicy(t *testing.T) {
	tests := []struct {
		name        string
		responses   []common.OCIOperationResponse
		shouldRetry bool
	}{
		{
			name: "testRetryPolicyWantRetry",
			responses: []common.OCIOperationResponse{
				mockOCIOperationResponseWithError(409, "IncorrectState"),
				mockOCIOperationResponseWithError(429, "TooManyRequests"),
				mockOCIOperationResponseWithError(500, "InternalServiceError"),
				mockOCIOperationResponseWithError(500, "OutOfCapacity"),
				mockOCIOperationResponseWithError(503, "ServiceUnavailable"),
			},
			shouldRetry: true,
		},
		{
			name: "testRetryPolicyNoRetry",
			responses: []common.OCIOperationResponse{
				mockOCIOperationResponseWithError(400, "CannotParseRequest"),
				mockOCIOperationResponseWithError(400, "InvalidParameter"),
				mockOCIOperationResponseWithError(400, "MissingParameter"),
				mockOCIOperationResponseWithError(400, "QuotaExceeded"),
				mockOCIOperationResponseWithError(400, "LimitExceeded"),
				mockOCIOperationResponseWithError(400, "RelatedResourceNotAuthorizedOrNotFound"),
				mockOCIOperationResponseWithError(400, "InsufficientServicePermissions"),
				mockOCIOperationResponseWithError(401, "NotAuthenticated"),
				mockOCIOperationResponseWithError(403, "SignUpRequired"),
				mockOCIOperationResponseWithError(403, "NotAllowed"),
				mockOCIOperationResponseWithError(403, "NotAuthorized"),
				mockOCIOperationResponseWithError(404, "NotFound"),
				mockOCIOperationResponseWithError(404, "InvalidParameter"),
				mockOCIOperationResponseWithError(404, "NotAuthorizedOrNotFound"),
				mockOCIOperationResponseWithError(405, "MethodNotAllowed"),
				mockOCIOperationResponseWithError(409, "NotAuthorizedOrResourceAlreadyExists"),
				mockOCIOperationResponseWithError(409, "InvalidatedRetryToken"),
				mockOCIOperationResponseWithError(412, "NoEtagMatch"),
				mockOCIOperationResponseWithError(413, "PayloadTooLarge"),
				mockOCIOperationResponseWithError(422, "UnprocessableEntity"),
				mockOCIOperationResponseWithError(431, "RequestHeaderFieldsTooLarge"),
				mockOCIOperationResponseWithError(501, "MethodNotImplemented"),
				mockOCIOperationResponseWithError(599, "Unknown 500 Error"),
				mockedOCIOperationResponse(200, 1),
			},
			shouldRetry: false,
		},
	}

	policy := newRetryPolicy()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, res := range tt.responses {
				assert.True(t, policy.ShouldRetryOperation(res) == tt.shouldRetry)
			}
		})
	}
}
