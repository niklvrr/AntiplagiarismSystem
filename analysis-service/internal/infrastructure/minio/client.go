package minio

import (
	"analysis-service/internal/config"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type Client struct {
	client *minio.Client
	bucket string
}

func NewClient(ctx context.Context, cfg *config.MinioConfig) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
	})
	if err != nil {
		return nil, handleMinioError(err)
	}

	exists, err := minioClient.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, handleMinioError(err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", cfg.Bucket)
	}

	return &Client{
		client: minioClient,
		bucket: cfg.Bucket,
	}, nil
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
