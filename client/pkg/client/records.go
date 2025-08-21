package client

import (
	"context"
	"fmt"

	"github.com/DanilNaum/secret-storage-client/internal/models"
	pb "github.com/DanilNaum/secret-storage-client/pkg/proto"
)

// CreateRecord creates a new record on the server
func (c *Client) CreateRecord(record *models.Record) (string, error) {
	if c.authToken == "" {
		return "", fmt.Errorf("not authenticated")
	}

	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	pbRecord, err := c.modelToProto(record)
	if err != nil {
		return "", fmt.Errorf("failed to convert record: %w", err)
	}

	req := &pb.CreateRecordRequest{
		AuthToken: c.authToken,
		Record:    pbRecord,
	}

	resp, err := c.recordClient.CreateRecord(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create record: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("create record failed: %s", resp.Message)
	}

	return resp.ServerId, nil
}

// GetRecord retrieves a record from the server
func (c *Client) GetRecord(serverID string) (*models.Record, error) {
	if c.authToken == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	req := &pb.GetRecordRequest{
		AuthToken:    c.authToken,
		ServerId:     serverID,
	}

	resp, err := c.recordClient.GetRecord(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get record failed: %s", resp.Message)
	}

	record, err := c.protoToModel(resp.Record)
	if err != nil {
		return nil, fmt.Errorf("failed to convert record: %w", err)
	}

	return record, nil
}

// UpdateRecord updates an existing record on the server
func (c *Client) UpdateRecord(record *models.Record) error {
	if c.authToken == "" {
		return fmt.Errorf("not authenticated")
	}

	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	pbRecord, err := c.modelToProto(record)
	if err != nil {
		return fmt.Errorf("failed to convert record: %w", err)
	}

	req := &pb.UpdateRecordRequest{
		AuthToken: c.authToken,
		Record:    pbRecord,
	}

	resp, err := c.recordClient.UpdateRecord(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("update record failed: %s", resp.Message)
	}

	return nil
}

// DeleteRecord deletes a record from the server
func (c *Client) DeleteRecord(serverID string) error {
	if c.authToken == "" {
		return fmt.Errorf("not authenticated")
	}

	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	req := &pb.DeleteRecordRequest{
		AuthToken: c.authToken,
		ServerId:  serverID,
	}

	resp, err := c.recordClient.DeleteRecord(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("delete record failed: %s", resp.Message)
	}

	return nil
}

// ListRecords retrieves all server records metadata
func (c *Client) ListRecords() ([]*models.ServerRecord, error) {
	if c.authToken == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	ctx, cancel := c.createContext(context.Background())
	defer cancel()

	req := &pb.ListRecordsRequest{
		AuthToken: c.authToken,
	}

	resp, err := c.recordClient.ListRecords(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("list records failed: %s", resp.Message)
	}

	var serverRecords []*models.ServerRecord
	for _, pbRecord := range resp.Records {
		serverRecord := &models.ServerRecord{
			ID:   pbRecord.Id,
			Name: pbRecord.Name,
			Type: models.RecordType(pbRecord.Type),
		}
		serverRecords = append(serverRecords, serverRecord)
	}

	return serverRecords, nil
}



// modelToProto converts models.Record to pb.Record (excludes local ID)
func (c *Client) modelToProto(record *models.Record) (*pb.Record, error) {
	pbRecord := &pb.Record{
		ServerId: record.ServerID,
		Name:     record.Name,
	}

	switch record.Type {
	case models.Credentials:
		pbRecord.Content = &pb.Record_Credentials{
			Credentials: &pb.CredentialsData{
				Username: record.Username,
				Password: record.Password,
			},
		}
	case models.TextData:
		pbRecord.Content = &pb.Record_TextData{
			TextData: &pb.TextData{
				TextContent: record.TextContent,
			},
		}
	case models.File:
		pbRecord.Content = &pb.Record_FileData{
			FileData: &pb.FileData{
				
			
			},
		}
	default:
		return nil, fmt.Errorf("unknown record type: %v", record.Type)
	}

	return pbRecord, nil
}

// protoToModel converts pb.Record to models.Record
func (c *Client) protoToModel(pbRecord *pb.Record) (*models.Record, error) {
	record := &models.Record{
		ServerID: pbRecord.ServerId,
		Name:     pbRecord.Name,
		IsServer: true, // Records from server are always server records
	}

	switch content := pbRecord.Content.(type) {
	case *pb.Record_Credentials:
		record.Type = models.Credentials
		record.Username = content.Credentials.Username
		record.Password = content.Credentials.Password
	case *pb.Record_TextData:
		record.Type = models.TextData
		record.TextContent = content.TextData.TextContent
	case *pb.Record_FileData:
		record.Type = models.File
	default:
		return nil, fmt.Errorf("unknown record content type")
	}

	return record, nil
}