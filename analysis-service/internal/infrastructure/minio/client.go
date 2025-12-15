package minio

import (
	"analysis-service/internal/config"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
	bucket string
}

func NewClient(ctx context.Context, cfg *config.MinioConfig) (*Client, error) {
	endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.InternalEndpoint, "http://"), "https://")

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
						return &Client{
							client: minioClient,
							bucket: cfg.Bucket,
						}, nil
					}
					return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
				}
			}
			return &Client{
				client: minioClient,
				bucket: cfg.Bucket,
			}, nil
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to check/create bucket after %d retries: %w", maxRetries, err)
}

func (c *Client) GetFile(ctx context.Context, objectKey string) ([]byte, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, handleMinioError(err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, handleMinioError(err)
	}

	return data, nil
}

func (c *Client) GetAllKeys(ctx context.Context) ([]string, error) {
	var files []string

	objectCh := c.client.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, handleMinioError(object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}
