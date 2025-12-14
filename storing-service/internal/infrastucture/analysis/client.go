package analysis

import (
	analysispb "analysis-service/pkg/api"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client analysispb.AnalysisServiceClient
}

func NewClient(ctx context.Context, endpoint string) (*Client, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: analysispb.NewAnalysisServiceClient(conn),
	}, nil
}

func (c *Client) AnalyseTask(ctx context.Context, taskId, objectKey string) (bool, error) {
	req := analysispb.AnalyzeTaskRequest{
		TaskId:    taskId,
		ObjectKey: objectKey,
	}

	resp, err := c.client.AnalyseTask(ctx, &req)
	if err != nil {
		return false, err
	}

	return resp.Status, nil
}

func (c *Client) GetReport(ctx context.Context, taskId string) (*analysispb.GetReportResponse, error) {
	req := analysispb.GetReportRequest{
		TaskId: taskId,
	}

	return c.client.GetReport(ctx, &req)
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
