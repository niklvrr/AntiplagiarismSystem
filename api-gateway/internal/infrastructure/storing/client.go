package storing

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	storingpb "storing-service/pkg/api"
)

type Client struct {
	conn   *grpc.ClientConn
	client storingpb.StoringServiceClient
}

func NewClient(ctx context.Context, endpoint string) (*Client, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to storing service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: storingpb.NewStoringServiceClient(conn),
	}, nil
}

func (c *Client) UploadTask(ctx context.Context, filename, uploadedBy string) (*storingpb.UploadTaskResponse, error) {
	req := &storingpb.UploadTaskRequest{
		Filename:   filename,
		UploadedBy: uploadedBy,
	}
	return c.client.UploadTask(ctx, req)
}

func (c *Client) GetTask(ctx context.Context, fileId string) (*storingpb.GetTaskResponse, error) {
	req := &storingpb.GetTaskRequest{
		FileId: fileId,
	}
	return c.client.GetTask(ctx, req)
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
