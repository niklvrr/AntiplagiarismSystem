package minio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"storing-service/internal/config"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	internalClient   *minio.Client
	externalEndpoint string
	bucket           string
}

func NewClient(ctx context.Context, cfg *config.MinioConfig) (*Client, error) {
	internalEndpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.InternalEndpoint, "http://"), "https://")
	externalEndpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.ExternalEndpoint, "http://"), "https://")

	var internalClient *minio.Client
	var err error

	maxRetries := 5
	retryDelay := 2 * time.Second

	// Создаем internal client для операций внутри Docker сети
	for i := 0; i < maxRetries; i++ {
		internalClient, err = minio.New(internalEndpoint, &minio.Options{
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
		return nil, fmt.Errorf("failed to create internal minio client after %d retries: %w", maxRetries, err)
	}

	// Проверяем bucket через internal client
	for i := 0; i < maxRetries; i++ {
		exists, err := internalClient.BucketExists(ctx, cfg.Bucket)
		if err == nil {
			if !exists {
				err = internalClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
				if err != nil {
					exists, checkErr := internalClient.BucketExists(ctx, cfg.Bucket)
					if checkErr == nil && exists {
						return &Client{
							internalClient:   internalClient,
							externalEndpoint: externalEndpoint,
							bucket:           cfg.Bucket,
						}, nil
					}
					return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
				}
			}
			return &Client{
				internalClient:   internalClient,
				externalEndpoint: externalEndpoint,
				bucket:           cfg.Bucket,
			}, nil
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to check/create bucket after %d retries: %w", maxRetries, err)
}

// PresignedPutObject генерирует presigned URL через internal client с Host header для external endpoint
func (c *Client) PresignedPutObject(ctx context.Context, objectKey string, expiry time.Duration) (*url.URL, error) {
	headers := make(http.Header)
	headers.Set("Host", c.externalEndpoint)

	presignedURL, err := c.internalClient.PresignHeader(ctx, http.MethodPut, c.bucket, objectKey, expiry, nil, headers)
	if err != nil {
		return nil, err
	}

	return c.replaceHost(presignedURL), nil
}

// PresignedGetObject генерирует presigned URL через internal client с Host header для external endpoint
func (c *Client) PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
	headers := make(http.Header)
	headers.Set("Host", c.externalEndpoint)

	presignedURL, err := c.internalClient.PresignHeader(ctx, http.MethodGet, c.bucket, objectKey, expiry, reqParams, headers)
	if err != nil {
		return nil, err
	}

	return c.replaceHost(presignedURL), nil
}

// replaceHost заменяет хост в URL на external endpoint
func (c *Client) replaceHost(u *url.URL) *url.URL {
	newURL := *u
	newURL.Host = c.externalEndpoint
	if !strings.HasPrefix(c.externalEndpoint, "http://") && !strings.HasPrefix(c.externalEndpoint, "https://") {
		if newURL.Scheme == "" {
			newURL.Scheme = "http"
		}
	}
	return &newURL
}

// StatObject использует internal client для проверки существования файла
func (c *Client) StatObject(ctx context.Context, objectKey string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return c.internalClient.StatObject(ctx, c.bucket, objectKey, opts)
}

// GetObject использует internal client для чтения файла
func (c *Client) GetObject(ctx context.Context, objectKey string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return c.internalClient.GetObject(ctx, c.bucket, objectKey, opts)
}
