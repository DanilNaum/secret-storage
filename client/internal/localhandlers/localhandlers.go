package localhandlers

import (
	"errors"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// StorageReader defines the interface for reading storage data.
type StorageReader interface {
	GetLocalRecords() []*models.Record
}

// StorageWriter defines the interface for writing storage data.
type StorageWriter interface {
	AddLocalRecord(record *models.Record)
	RemoveLocalRecord(id string) *models.Record
	FindLocalRecordByServerID(serverID string) *models.Record
	UpdateLocalRecordServerID(localID, serverID string) bool
}

// StorageAdapter combines reader and writer interfaces.
type StorageAdapter interface {
	StorageReader
	StorageWriter
}

// LocalHandlers defines all event handlers for local record operations.
type LocalHandlers struct {
	storage StorageReader
	writer  StorageWriter
}

// NewLocalHandlers creates local handlers with storage adapter.
func NewLocalHandlers(storage StorageAdapter) *LocalHandlers {
	return &LocalHandlers{
		storage: storage,
		writer:  storage,
	}
}

// CreateRecord creates a new local record.
func (h *LocalHandlers) CreateRecord(record *models.Record) error {
	
	h.writer.AddLocalRecord(record)
	return nil
}

// DeleteLocalRecord deletes a local record.
func (h *LocalHandlers) DeleteLocalRecord(record *models.Record) error {
	removedRecord := h.writer.RemoveLocalRecord(record.ID)
	if removedRecord == nil {
		return errors.New("local record not found")
	}
	return nil
}

// LocalRecordChange updates an existing local record.
func (h *LocalHandlers) LocalRecordChange(record *models.Record) error {
	// В новой реализации с картами мы не можем напрямую обновить запись,
	// поэтому удалим старую и добавим новую
	if h.writer.RemoveLocalRecord(record.ID) == nil {
		return errors.New("local record not found")
	}

	// Добавляем обновленную запись
	h.writer.AddLocalRecord(record)
	return nil
}

// LocalRecordOpen loads a local record by ID.
func (h *LocalHandlers) LocalRecordOpen(id string) (*models.Record, error) {
	// Получаем все записи
	localRecords := h.storage.GetLocalRecords()

	// Ищем запись по ID
	for _, r := range localRecords {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errors.New("local record not found")
}

// LocalRecordMovedToServer uploads a local record to the server.
func (h *LocalHandlers) LocalRecordMovedToServer(record *models.Record, serverID string) error {
	h.writer.UpdateLocalRecordServerID(record.ID, serverID)
	return nil
}

// ServerRecordMovedToLocal downloads a server record to local storage.
func (h *LocalHandlers) ServerRecordMovedToLocal(record *models.Record) error {
	if existingRecord := h.writer.FindLocalRecordByServerID(record.ServerID); existingRecord != nil {
		// Обновляем существующую запись
		existingRecord.Name = record.Name
		existingRecord.Type = record.Type
		existingRecord.Username = record.Username
		existingRecord.Password = record.Password
		existingRecord.TextContent = record.TextContent
		existingRecord.FilePath = record.FilePath
	} else {
		// Создаем новую запись
		newRecord := record.Clone()
		newRecord.ServerID = record.ServerID
		newRecord.ID = "" // ID будет сгенерирован при добавлении
		newRecord.IsServer = false // Это локальная копия серверной записи
		h.writer.AddLocalRecord(newRecord)
	}

	return nil
}