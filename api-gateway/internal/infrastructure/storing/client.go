package storing

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"go.uber.org/zap"
	storingpb "storing-service/pkg/api"
)

type Client struct {
	conn   *grpc.ClientConn
	client storingpb.StoringServiceClient
	logger *zap.Logger
}

func NewClient(ctx context.Context, endpoint string, logger *zap.Logger) (*Client, error) {
	logger.Info("connecting to storing service", zap.String("endpoint", endpoint))
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to storing service", zap.String("endpoint", endpoint), zap.Error(err))
		return nil, fmt.Errorf("failed to connect to storing service: %w", err)
	}

	logger.Info("connected to storing service", zap.String("endpoint", endpoint))
	return &Client{
		conn:   conn,
		client: storingpb.NewStoringServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *Client) UploadTask(ctx context.Context, filename, uploadedBy string) (*storingpb.UploadTaskResponse, error) {
	c.logger.Debug("calling storing service UploadTask",
		zap.String("filename", filename),
		zap.String("uploaded_by", uploadedBy))

	res, err := c.client.UploadTask(ctx, &storingpb.UploadTaskRequest{
		Filename:   filename,
		UploadedBy: uploadedBy,
	})

	if err != nil {
		c.logger.Error("storing service UploadTask failed",
			zap.String("filename", filename),
			zap.Error(err))
		return nil, err
	}

	c.logger.Debug("storing service UploadTask success",
		zap.String("file_id", res.FileId))
	return res, nil
}

func (c *Client) GetTask(ctx context.Context, fileId string) (*storingpb.GetTaskResponse, error) {
	c.logger.Debug("calling storing service GetTask", zap.String("file_id", fileId))

	res, err := c.client.GetTask(ctx, &storingpb.GetTaskRequest{
		FileId: fileId,
	})

	if err != nil {
		c.logger.Error("storing service GetTask failed",
			zap.String("file_id", fileId),
			zap.Error(err))
		return nil, err
	}

	c.logger.Debug("storing service GetTask success", zap.String("file_id", fileId))
	return res, nil
}

func (c *Client) GetFileContent(ctx context.Context, fileId string) (*storingpb.GetFileContentResponse, error) {
	c.logger.Debug("calling storing service GetFileContent", zap.String("file_id", fileId))

	res, err := c.client.GetFileContent(ctx, &storingpb.GetFileContentRequest{
		FileId: fileId,
	})

	if err != nil {
		c.logger.Error("storing service GetFileContent failed",
			zap.String("file_id", fileId),
			zap.Error(err))
		return nil, err
	}

	c.logger.Debug("storing service GetFileContent success",
		zap.String("file_id", fileId),
		zap.Int("content_size", len(res.Content)))
	return res, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
