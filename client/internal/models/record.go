package models

// RecordType represents the type of a record in the storage system.
type RecordType int

const (
	// Credentials represents a record containing username/password pairs.
	Credentials RecordType = iota
	// TextData represents a record containing plain text content.
	TextData
	// File represents a record containing file path information.
	File
)

// Record represents a single record in the storage system.
type Record struct {
	ID          string
	ServerID    string
	Type        RecordType
	Name        string
	Username    string
	Password    string
	TextContent string
	FilePath    string
	IsServer    bool
}

// ServerRecord represents server record metadata without full content.
type ServerRecord struct {
	ID   string
	Name string
	Type RecordType
}

// String returns the string representation of RecordType.
func (rt RecordType) String() string {
	switch rt {
	case Credentials:
		return "Credentials"
	case TextData:
		return "Text"
	case File:
		return "File"
	default:
		return "Unknown"
	}
}

// Clone creates a deep copy of the record.
func (r *Record) Clone() *Record {
	return &Record{
		ID:          r.ID,
		ServerID:    r.ServerID,
		Type:        r.Type,
		Name:        r.Name,
		Username:    r.Username,
		Password:    r.Password,
		TextContent: r.TextContent,
		FilePath:    r.FilePath,
		IsServer:    r.IsServer,
	}
}

// HasServerID checks if the record is synchronized with server.
func (r *Record) HasServerID() bool {
	return r.ServerID != ""
}