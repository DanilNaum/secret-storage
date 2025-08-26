package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/DanilNaum/secret-storage-client/pkg/proto"
)

// Client represents the gRPC client for secret storage service
type Client struct {
	conn         *grpc.ClientConn
	authClient   pb.AuthServiceClient
	recordClient pb.RecordServiceClient
	authToken    string
	salt         string
	serverURL    string
	timeout      time.Duration
}

// ClientConfig holds configuration for the client
type ClientConfig struct {
	ServerURL  string
	Timeout    time.Duration
	UseTLS     bool
	SkipVerify bool
}

// NewClient creates a new gRPC client instance
func NewClient(config ClientConfig) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	var opts []grpc.DialOption

	if config.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: config.SkipVerify,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(config.ServerURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &Client{
		conn:         conn,
		authClient:   pb.NewAuthServiceClient(conn),
		recordClient: pb.NewRecordServiceClient(conn),
		serverURL:    config.ServerURL,
		timeout:      config.Timeout,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsAuthenticated returns whether the client has a valid auth token
func (c *Client) IsAuthenticated() bool {
	return c.authToken != ""
}

// GetAuthToken returns the current auth token
func (c *Client) GetAuthToken() string {
	return c.authToken
}

// SetAuthToken sets the auth token manually
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// createContext creates a context with timeout
func (c *Client) createContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}
