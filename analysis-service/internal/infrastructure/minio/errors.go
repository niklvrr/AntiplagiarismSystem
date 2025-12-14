package minio

import (
	"analysis-service/internal/errdefs"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func handleMinioError(err error) error {
	if err == nil {
		return nil
	}

	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) {
		switch minioErr.Code {
		case "NoSuchKey", "NoSuchBucket":
			return fmt.Errorf("%w: %v", errdefs.ErrNotFound, err)
		case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch":
			return fmt.Errorf("%w: %v", errdefs.ErrUnavailable, err)
		case "InvalidBucketName", "InvalidObjectName":
			return fmt.Errorf("%w: %v", errdefs.ErrInvalidArgument, err)
		default:
			return fmt.Errorf("%w: %v", errdefs.ErrUnavailable, err)
		}
	}

	var credErr credentials.ErrorResponse
	if errors.As(err, &credErr) {
		return fmt.Errorf("%w: invalid credentials: %v", errdefs.ErrUnavailable, err)
	}

	if isNetworkError(err) {
		return fmt.Errorf("%w: minio connection failed: %v", errdefs.ErrUnavailable, err)
	}

	if isReadError(err) {
		return fmt.Errorf("%w: %v", errdefs.ErrInvalidArgument, err)
	}

	return fmt.Errorf("%w: %v", errdefs.ErrUnavailable, err)
}

func isNetworkError(err error) bool {
	errStr := err.Error()
	networkErrors := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"timeout",
		"deadline exceeded",
		"network is unreachable",
		"dial tcp",
	}

	for _, netErr := range networkErrors {
		if contains(errStr, netErr) {
			return true
		}
	}
	return false
}

func isReadError(err error) bool {
	errStr := err.Error()
	readErrors := []string{
		"unexpected EOF",
		"read:",
		"io:",
	}

	for _, readErr := range readErrors {
		if contains(errStr, readErr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
