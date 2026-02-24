package ociclient

import (
	"context"
	"io"
	"time"

	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// Gets the namespace of the current caller
func (c *client) GetNamespace(ctx context.Context) (string, error) {
	req := objectstorage.GetNamespaceRequest{
		RequestMetadata: c.requestMetadata,
	}

	resp, err := c.objectstorage.GetNamespace(ctx, req)
	incRequestCounter(err, getVerb, objectStorageResource)
	if err != nil {
		return "", err
	}

	return *resp.Value, nil
}

func (c *client) ListObjects(ctx context.Context, namespace, bucketName string) ([]objectstorage.ObjectSummary, error) {
	req := objectstorage.ListObjectsRequest{
		NamespaceName:   &namespace,
		BucketName:      &bucketName,
		RequestMetadata: c.requestMetadata,
		Limit:           &listLimit,
	}

	var result []objectstorage.ObjectSummary
	for {
		resp, err := c.objectstorage.ListObjects(ctx, req)
		incRequestCounter(err, listVerb, objectStorageResource)
		if err != nil {
			return nil, NewOCIClientError(resp.OpcRequestId, err)
		}

		result = append(result, resp.Objects...)
		if resp.NextStartWith == nil {
			break
		}

		req.Start = resp.NextStartWith
		time.Sleep(50 * time.Millisecond)
	}

	return result, nil

}

func (c *client) DeleteObject(ctx context.Context, namespace, bucketName, objectName string) (string, error) {
	req := objectstorage.DeleteObjectRequest{
		NamespaceName:   &namespace,
		BucketName:      &bucketName,
		ObjectName:      &objectName,
		RequestMetadata: c.requestMetadata,
	}
	resp, err := c.objectstorage.DeleteObject(ctx, req)
	incRequestCounter(err, deleteVerb, objectStorageResource)
	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}
	return *resp.OpcRequestId, err
}

func (c *client) DeleteBucket(ctx context.Context, namespace, bucketName string) (string, error) {
	req := objectstorage.DeleteBucketRequest{
		NamespaceName:   &namespace,
		BucketName:      &bucketName,
		RequestMetadata: c.requestMetadata,
	}

	resp, err := c.objectstorage.DeleteBucket(ctx, req)
	incRequestCounter(err, deleteVerb, objectStorageResource)

	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcRequestId, nil
}

func (c *client) GetObject(ctx context.Context, namespace string, bucketName string, objectName string) (objectstorage.GetObjectResponse, error) {
	req := objectstorage.GetObjectRequest{
		NamespaceName:   &namespace,
		BucketName:      &bucketName,
		ObjectName:      &objectName,
		RequestMetadata: c.requestMetadata,
	}

	resp, err := c.objectstorage.GetObject(ctx, req)
	incRequestCounter(err, getVerb, objectStorageResource)

	if err != nil {
		return objectstorage.GetObjectResponse{}, NewOCIClientError(resp.OpcRequestId, err)
	}

	return resp, nil
}

func (c *client) PutObject(ctx context.Context, namespace string, bucketName string,
	objectName string, contentLength int64, object io.ReadCloser) (string, error) {

	req := objectstorage.PutObjectRequest{
		NamespaceName:   &namespace,
		BucketName:      &bucketName,
		ObjectName:      &objectName,
		ContentLength:   &contentLength,
		PutObjectBody:   object,
		RequestMetadata: c.requestMetadata,
	}

	resp, err := c.objectstorage.PutObject(ctx, req)

	incRequestCounter(err, updateVerb, objectStorageResource)

	if err != nil {
		return "", NewOCIClientError(resp.OpcRequestId, err)
	}

	return *resp.OpcRequestId, nil
}
