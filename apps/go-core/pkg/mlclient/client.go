// apps/go-core/pkg/mlclient/client.go
package mlclient

import (
	"context"
	"time"

	mlv1 "github.com/bimal009/Zovly/gen/ml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
	chat mlv1.ChatServiceClient
}

// New dials the py-ml gRPC server, e.g. "localhost:50051".
func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, chat: mlv1.NewChatServiceClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) Chat(ctx context.Context, message string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.chat.Send(ctx, &mlv1.ChatRequest{Message: message})
	if err != nil {
		return "", err
	}
	return resp.Reply, nil
}
