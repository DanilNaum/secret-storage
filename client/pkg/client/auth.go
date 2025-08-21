package client

import (
	"context"
	"fmt"

	pb "github.com/DanilNaum/secret-storage-client/pkg/proto"
)

// Register performs user registration with the server
func (c *Client) Register(username, password string) error {
	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	req := &pb.RegisterRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.authClient.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	c.authToken = resp.AuthToken
	c.salt = resp.Salt
	return nil
}

// Authenticate performs user authentication with the server
func (c *Client) Authenticate(username, password string) error {
	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	req := &pb.AuthenticateRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.authClient.Authenticate(ctx, req)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("authentication failed: %s", resp.Message)
	}

	c.authToken = resp.AuthToken
	c.salt = resp.Salt
	return nil
}

// Logout performs user logout
func (c *Client) Logout() error {
	if c.authToken == "" {
		return nil // Already logged out
	}

	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	req := &pb.LogoutRequest{
		AuthToken: c.authToken,
	}

	resp, err := c.authClient.Logout(ctx, req)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("logout failed: %s", resp.Message)
	}

	c.authToken = ""
	c.salt = ""
	return nil
}
