package storage

import (
	"fmt"
	"github.com/elliotchance/orderedmap/v2"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// Storage manages local and server records in memory using maps for efficient access.
type Storage struct {
	localRecords  *orderedmap.OrderedMap[string, *models.Record]       // Ключ: ID локальной записи
	serverRecords *orderedmap.OrderedMap[string, *models.ServerRecord] // Ключ: ID серверной записи
}

// NewStorage creates a new storage instance with predefined test data.
func NewStorage() *Storage {
	return &Storage{
		localRecords:  orderedmap.NewOrderedMap[string, *models.Record](),
		serverRecords: orderedmap.NewOrderedMap[string, *models.ServerRecord](),
	}
}

// GetLocalRecords returns all local records.
func (s *Storage) GetLocalRecords() []*models.Record {
	records := make([]*models.Record, 0, s.localRecords.Len())
	// s.localRecords.
	for record := s.localRecords.Front(); record != nil; record = record.Next() {
		records = append(records, record.Value)
	}
	return records
}

// GetServerRecords returns all server record metadata.
func (s *Storage) GetServerRecords() []*models.ServerRecord {
	records := make([]*models.ServerRecord, 0, s.serverRecords.Len())
	for record := s.serverRecords.Front(); record != nil; record = record.Next() {
		records = append(records, record.Value)
	}
	return records
}

// AddLocalRecord adds a new record to local storage.
func (s *Storage) AddLocalRecord(record *models.Record) {
	if record.ID == "" {
		// Генерируем новый ID, если он не указан
		for {
			record.ID = fmt.Sprintf("local_%d", (s.localRecords.Len())+1)
			if _, exists := s.localRecords.Get(record.ID); !exists {
				break
			}
		}
	}
	record.IsServer = false
	s.localRecords.Set(record.ID, record)
}

// AddServerRecord adds server record metadata to storage.
func (s *Storage) AddServerRecord(serverRecord *models.ServerRecord) {
	if serverRecord.ID == "" {
		// Генерируем новый ID, если он не указан
		for {
			serverRecord.ID = fmt.Sprintf("server_%d", (s.serverRecords.Len())+1)
			if _, exists := s.serverRecords.Get(serverRecord.ID); !exists {
				break
			}
		}
	}
	s.serverRecords.Set(serverRecord.ID, serverRecord)
}

func (s *Storage) SetServerRecords(serverRecords []*models.ServerRecord) {
	newServerRecords := orderedmap.NewOrderedMap[string, *models.ServerRecord]()
	for _, record := range serverRecords {
		newServerRecords.Set(record.ID, record)
	}
	s.serverRecords = newServerRecords
}

// RemoveLocalRecord removes a local record by ID.
func (s *Storage) RemoveLocalRecord(id string) *models.Record {
	if id == "" {
		return nil
	}
	record, exists := s.localRecords.Get(id)
	if !exists {
		return nil
	}
	s.localRecords.Delete(id)
	return record
}

// RemoveServerRecord removes a server record metadata by server ID.
func (s *Storage) RemoveServerRecord(serverID string) *models.ServerRecord {
	if serverID == "" {
		return nil
	}
	record, exists := s.serverRecords.Get(serverID)
	if !exists {
		return nil
	}
	s.serverRecords.Delete(serverID)
	return record
}

// UpdateLocalRecordServerID updates the server ID of a local record.
func (s *Storage) UpdateLocalRecordServerID(localID, serverID string) bool {
	if localID == "" {
		return false
	}

	record, exists := s.localRecords.Get(localID)
	if !exists {
		return false
	}

	record.ServerID = serverID
	return true
}

// FindLocalRecordByServerID finds a local record by its server ID.
func (s *Storage) FindLocalRecordByServerID(serverID string) *models.Record {
	if serverID == "" {
		return nil
	}

	for record:= s.localRecords.Front(); record != nil; record = record.Next() {
		
	
		if record.Value.ServerID == serverID {
			return record.Value
		}
	}
	return nil
}
