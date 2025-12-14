package minio

import (
	"context"
	"fmt"
	"storing-service/internal/config"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClient(ctx context.Context, cfg *config.MinioConfig) (*minio.Client, error) {
	endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "http://"), "https://")

	var minioClient *minio.Client
	var err error

	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		minioClient, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: false,
		})
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client after %d retries: %w", maxRetries, err)
	}

	for i := 0; i < maxRetries; i++ {
		exists, err := minioClient.BucketExists(ctx, cfg.Bucket)
		if err == nil {
			if !exists {
				err = minioClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
				if err != nil {
					exists, checkErr := minioClient.BucketExists(ctx, cfg.Bucket)
					if checkErr == nil && exists {
						return minioClient, nil
					}
					return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
				}
			}
			return minioClient, nil
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to check/create bucket after %d retries: %w", maxRetries, err)
}
