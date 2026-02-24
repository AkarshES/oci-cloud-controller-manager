package ociclient

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/pkg/errors"
)

const missingOPCRequestID = "missingno"

func NewOCIClientError(opcRequestID *string, e error) OCIClientError {
	opcReqID := missingOPCRequestID
	if opcRequestID != nil {
		opcReqID = *opcRequestID
	}
	return OCIClientError{
		err:          e,
		opcRequestID: opcReqID,
	}
}

type OCIClientError struct {
	err          error
	opcRequestID string
}

// Cause implements the errors.Causer interface and returns the underlying error
func (o OCIClientError) Cause() error { return o.err }

// Error returns a string comprised of the underlying error wrapped with the opc request ID
// This forces clients to have the opc request ID in all log messages
func (o OCIClientError) Error() string {
	return fmt.Sprintf("%s opcRequestID %s", o.err, o.opcRequestID)
}

// GetOpcRequestId returns just the opc request id.
// This method is available for caller to have the ID by itself for use such as logging it with it's own field
func (o OCIClientError) GetOpcRequestId() string {
	return o.opcRequestID
}

// IsNotFound returns true if the given error indicates that a resource could
// not be found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)
	if err == ErrNotFound {
		return true
	}

	serviceErr, ok := common.IsServiceError(err)
	return ok && serviceErr.GetHTTPStatusCode() == http.StatusNotFound
}

// IsRetryable returns true if the given error is retriable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)
	serviceErr, ok := common.IsServiceError(err)
	if !ok {
		return false
	}

	switch serviceErr.GetHTTPStatusCode() {
	case http.StatusTooManyRequests, http.StatusGatewayTimeout, http.StatusBadGateway:
		return true
	default:
		return false
	}
}

func newRetryPolicy() *common.RetryPolicy {
	return NewRetryPolicyWithMaxAttempts(uint(2))
}

// NewRetryPolicyWithMaxAttempts returns a RetryPolicy with the specified max retryAttempts
func NewRetryPolicyWithMaxAttempts(retryAttempts uint) *common.RetryPolicy {
	isRetryableOperation := func(r common.OCIOperationResponse) bool {
		return IsRetryable(r.Error)
	}

	nextDuration := func(r common.OCIOperationResponse) time.Duration {
		// you might want wait longer for next retry when your previous one failed
		// this function will return the duration as:
		// 1s, 2s, 4s, 8s, 16s, 32s, 64s etc...
		return time.Duration(math.Pow(float64(2), float64(r.AttemptNumber-1))) * time.Second
	}

	policy := common.NewRetryPolicy(
		retryAttempts, isRetryableOperation, nextDuration,
	)
	return &policy
}
