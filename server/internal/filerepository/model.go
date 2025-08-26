package filerepository

type FileMetadata struct {
	RecordID    string
	*FileInfo
}

type FileInfo struct {
	Filename    string
	FileSize    int64
	ContentType string
}