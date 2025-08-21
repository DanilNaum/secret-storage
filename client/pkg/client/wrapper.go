package client

import (
	"fmt"
	"time"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// ServerClient wraps the gRPC client to match the existing ServerHandlers interface
type ServerClient struct {
	client         *Client
	masterPassword string
}

// NewServerClient creates a new server client wrapper
func NewServerClient(serverURL string, useTLS bool) (*ServerClient, error) {
	config := ClientConfig{
		ServerURL:  serverURL,
		Timeout:    30 * time.Second,
		UseTLS:     useTLS,
		SkipVerify: true, // For development
	}

	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ServerClient{
		client: client,
	}, nil
}

// Close closes the underlying gRPC connection
func (sc *ServerClient) Close() error {
	return sc.client.Close()
}

// Authenticate performs user authentication with the server
func (sc *ServerClient) Authenticate(username, password string) error {
	return sc.client.Authenticate(username, password)
}

// Register performs user registration with the server
func (sc *ServerClient) Register(username, password string) error {
	return sc.client.Register(username, password)
}

// SetMasterPassword sets the master password for encryption
func (sc *ServerClient) SetMasterPassword(password string) {
	sc.masterPassword = password
}

// Logout performs user logout
func (sc *ServerClient) Logout() error {
	err := sc.client.Logout()
	if err == nil {
		sc.masterPassword = ""
	}
	return err
}

// IsAuthenticated returns the authentication status
func (sc *ServerClient) IsAuthenticated() bool {
	return sc.client.IsAuthenticated()
}

// DeleteServerRecord deletes a server record
func (sc *ServerClient) DeleteServerRecord(record *models.Record) error {
	if record.ServerID == "" {
		return fmt.Errorf("record has no server ID")
	}
	return sc.client.DeleteRecord(record.ServerID)
}

// ServerRecordChange updates an existing server record
func (sc *ServerClient) ServerRecordChange(record *models.Record) error {
	if record.ServerID == "" {
		return fmt.Errorf("record has no server ID")
	}
	return sc.client.UpdateRecord(record)
}

// ServerRecordOpen loads a server record by ID
func (sc *ServerClient) ServerRecordOpen(serverID string) (*models.Record, error) {
	return sc.client.GetRecord(serverID)
}

// Sync performs full synchronization between local and server storage

// LocalRecordMovedToServer creates a new server record from a local record
func (sc *ServerClient) LocalRecordMovedToServer(record *models.Record) (string, error) {
	serverID, err := sc.client.CreateRecord(record)
	if err != nil {
		return "", err
	}

	// If it's a file record, upload the file
	if record.Type == models.File && record.FilePath != "" {
		go sc.uploadFile(serverID, record.FilePath)
	}

	return serverID, nil
}

// UploadFile uploads a file for a record
func (sc *ServerClient) uploadFile(serverID, filePath string) {
	err := sc.client.UploadFile(serverID, filePath)
	if err != nil {
		// Try to clean up the created record
		sc.client.DeleteRecord(serverID)
		return
	}

}

func (sc *ServerClient) ListRecords() ([]*models.ServerRecord, error) {
	return sc.client.ListRecords()
}

func (sc *ServerClient) DownloadFile(recordID, savePath string) (chan int, chan error) {
	return sc.client.DownloadFile(recordID, savePath)
}
