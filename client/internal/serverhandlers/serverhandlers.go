package serverhandlers

import (
	"fmt"
	"time"

	"github.com/DanilNaum/secret-storage-client/internal/models"
	"github.com/DanilNaum/secret-storage-client/pkg/client"
)

// StorageReader defines the interface for reading storage data.
type StorageReader interface {
	GetLocalRecords() []*models.Record
	GetServerRecords() []*models.ServerRecord
}

// StorageWriter defines the interface for writing storage data.
type StorageWriter interface {
	AddLocalRecord(record *models.Record)
	AddServerRecord(serverRecord *models.ServerRecord)
	RemoveServerRecord(serverID string) *models.ServerRecord
	SetServerRecords(records []*models.ServerRecord)
}

// CacheManager defines the interface for cache operations.
type CacheManager interface {
	Get(id string) (*models.Record, bool, time.Time)
	Set(id string, record *models.Record)
	Remove(id string)
	IsExpired(id string) bool
}

// StorageAdapter combines reader and writer interfaces.
type StorageAdapter interface {
	StorageReader
	StorageWriter
}

// ServerHandlers defines all event handlers for server record operations.
type ServerHandlers struct {
	*client.ServerClient
	cache   CacheManager
	storage StorageAdapter
}

// NewServerHandlers creates server handlers with storage and cache adapters.
func NewServerHandlers(storage StorageAdapter, cache CacheManager, serverURL string) (*ServerHandlers, error) {
	// Create gRPC client wrapper
	grpcClient, err := client.NewServerClient(serverURL, false) // Use TLS by default
	if err != nil {
		// In case of error, create a placeholder client that will show error messages
		fmt.Printf("failed to create gRPC client: %v. Using placeholder client.", err)

	}

	return &ServerHandlers{
		ServerClient: grpcClient,
		cache:        cache,
		storage:      storage,
	}, nil
}

func (h *ServerHandlers) DeleteServerRecord(record *models.Record) error {
	err := h.ServerClient.DeleteServerRecord(record)
	if err != nil {
		return err
	}
	h.cache.Remove(record.ServerID)

	// Удаляем запись из хранилища по ID
	h.storage.RemoveServerRecord(record.ServerID)

	return nil
}

func (h *ServerHandlers) LocalRecordMovedToServer(record *models.Record) (string, error) {
	serverID, err := h.ServerClient.LocalRecordMovedToServer(record)
	if err != nil {
		return "", err
	}
	serverRecord := record.Clone()
	serverRecord.IsServer = true
	serverRecord.ServerID = serverID
	h.cache.Set(serverID, serverRecord)
	h.storage.AddServerRecord(&models.ServerRecord{
		ID:   serverID,
		Name: serverRecord.Name,
		Type: serverRecord.Type,
	})

	return serverID, nil
}

// GetCachedRecord retrieves a record from cache with metadata.
func (h *ServerHandlers) GetCachedRecord(id string) (*models.Record, bool, time.Time, bool) {
	record, found, cachedAt := h.cache.Get(id)
	expired := h.cache.IsExpired(id)
	return record, found, cachedAt, expired
}

// SetCachedRecord stores a record in cache.
func (h *ServerHandlers) SetCachedRecord(id string, record *models.Record) {
	h.cache.Set(id, record)
}

// ServerRecordChange updates an existing server record.
func (h *ServerHandlers) ServerRecordChange(record *models.Record) error {

	err := h.ServerClient.ServerRecordChange(record)
	if err != nil {
		return err
	}
	h.cache.Set(record.ServerID, record)

	return nil
}

// ServerRecordOpen loads a server record by ID.
func (h *ServerHandlers) ServerRecordOpen(serverID string) (*models.Record, error) {

	if cachedRecord, found, _ := h.cache.Get(serverID); found {
		return cachedRecord, nil
	}
	record, err := h.ServerClient.ServerRecordOpen(serverID)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// RefreshServerRecord forces a refresh of a server record from the server.
func (h *ServerHandlers) RefreshServerRecord(serverID string) (*models.Record, error) {
	record, err := h.ServerClient.ServerRecordOpen(serverID)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// UpdateServerRecordList is a method of the ServerHandlers struct used to update the server record list
// This method fetches server records from ServerClient and stores them in storage
func (h *ServerHandlers) UpdateServerRecordList() error {
	serverRecords, err := h.ServerClient.ListRecords()
	if err != nil {
		return err
	}
	h.storage.SetServerRecords(serverRecords)
	return nil
}

func (h *ServerHandlers) DownloadFile(recordID string, savePath string) (chan int, chan error) {
	return h.ServerClient.DownloadFile(recordID, savePath)
}