package analysis

import (
	analysispb "analysis-service/pkg/api"
	"context"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client analysispb.AnalysisServiceClient
	logger *zap.Logger
}

func NewClient(ctx context.Context, endpoint string, logger *zap.Logger) (*Client, error) {
	logger.Info("connecting to analysis service", zap.String("endpoint", endpoint))
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to analysis service", zap.String("endpoint", endpoint), zap.Error(err))
		return nil, fmt.Errorf("failed to connect to analysis service: %w", err)
	}

	logger.Info("connected to analysis service", zap.String("endpoint", endpoint))
	return &Client{
		conn:   conn,
		client: analysispb.NewAnalysisServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *Client) AnalyseTask(ctx context.Context, taskId, filename string) (*analysispb.AnalyseTaskResponse, error) {
	objectKey := makeObjectKey(taskId, filename)
	c.logger.Debug("calling analysis service AnalyseTask",
		zap.String("task_id", taskId),
		zap.String("object_key", objectKey))

	res, err := c.client.AnalyseTask(ctx, &analysispb.AnalyzeTaskRequest{
		TaskId:    taskId,
		ObjectKey: objectKey,
	})

	if err != nil {
		c.logger.Error("analysis service AnalyseTask failed",
			zap.String("task_id", taskId),
			zap.Error(err))
		return nil, err
	}

	c.logger.Debug("analysis service AnalyseTask success",
		zap.String("task_id", taskId),
		zap.Bool("status", res.Status))
	return res, nil
}

func (c *Client) GetReport(ctx context.Context, taskId string) (*analysispb.GetReportResponse, error) {
	c.logger.Debug("calling analysis service GetReport", zap.String("task_id", taskId))

	res, err := c.client.GetReport(ctx, &analysispb.GetReportRequest{
		TaskId: taskId,
	})

	if err != nil {
		c.logger.Error("analysis service GetReport failed",
			zap.String("task_id", taskId),
			zap.Error(err))
		return nil, err
	}

	c.logger.Debug("analysis service GetReport success",
		zap.String("task_id", taskId),
		zap.Bool("is_plagiarism", res.IsPlagiarism))
	return res, nil
}

func (c *Client) GenerateWordCloud(ctx context.Context, fileContent []byte) (*analysispb.GenerateWordCloudResponse, error) {
	c.logger.Debug("calling analysis service GenerateWordCloud",
		zap.Int("content_size", len(fileContent)))

	res, err := c.client.GenerateWordCloud(ctx, &analysispb.GenerateWordCloudRequest{
		FileContent: fileContent,
	})

	if err != nil {
		c.logger.Error("analysis service GenerateWordCloud failed", zap.Error(err))
		return nil, err
	}

	c.logger.Debug("analysis service GenerateWordCloud success",
		zap.String("image_url", res.ImageUrl))
	return res, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func makeObjectKey(id, filename string) string {
	ext := filepath.Ext(filename)
	return fmt.Sprintf("%s%s", id, ext)
}
