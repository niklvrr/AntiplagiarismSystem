package analysis

import (
	analysispb "analysis-service/pkg/api"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"path/filepath"
)

type Client struct {
	conn   *grpc.ClientConn
	client analysispb.AnalysisServiceClient
}

func NewClient(ctx context.Context, endpoint string) (*Client, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to analysis service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: analysispb.NewAnalysisServiceClient(conn),
	}, nil
}

func (c *Client) AnalyseTask(ctx context.Context, taskId, filename string) (*analysispb.AnalyseTaskResponse, error) {
	req := &analysispb.AnalyzeTaskRequest{
		TaskId:    taskId,
		ObjectKey: makeObjectKey(taskId, filename),
	}
	return c.client.AnalyseTask(ctx, req)
}

func (c *Client) GetReport(ctx context.Context, taskId string) (*analysispb.GetReportResponse, error) {
	req := &analysispb.GetReportRequest{
		TaskId: taskId,
	}
	return c.client.GetReport(ctx, req)
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
