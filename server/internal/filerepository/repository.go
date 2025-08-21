package filerepository

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type fileRepository struct {
	storagePath string
	mu          sync.RWMutex
	storage     map[string]*FileInfo
}

func NewFileRepository(storagePath string) *fileRepository {
	// Создаем директорию для хранения если не существует
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create storage directory: %v", err))
	}

	return &fileRepository{
		storagePath: storagePath,
		storage:     make(map[string]*FileInfo),
	}
}

func (r *fileRepository) getFilePath(fileID string) string {
	return filepath.Join(r.storagePath, fileID)
}

func (r *fileRepository) SaveFile(metadata *FileMetadata, chunks <-chan []byte, errChan <-chan error) error {
	filePath := r.getFilePath(metadata.RecordID)
	r.storage[metadata.RecordID] = metadata.FileInfo

	// Создаем файл
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	var totalSize int64

	// Записываем чанки в файл
LOOP:
	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				break LOOP
			}
			if _, err := file.Write(chunk); err != nil {
				return fmt.Errorf("failed to write chunk: %w", err)
			}
			totalSize += int64(len(chunk))
		case err := <-errChan:
			return err
		}
	}

	// Проверяем размер файла
	if totalSize != metadata.FileSize {
		os.Remove(filePath) // Удаляем неполный файл
		return fmt.Errorf("file size mismatch: expected %d, got %d", metadata.FileSize, totalSize)
	}

	return nil
}

func (r *fileRepository) GetFile(fileID string) (*FileInfo, <-chan []byte, error) {
	filePath := r.getFilePath(fileID)

	// Проверяем существование файла
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("file not found: %w", err)
	}

	fileInfo := &FileInfo{
		Filename:    fileID, // или можно извлечь оригинальное имя из метаданных
		FileSize:    info.Size(),
		ContentType: "application/octet-stream",
	}
	if info, ok := r.storage[fileID]; ok {
		fileInfo = info
	}

	// Создаем канал для потоковой передачи чанков
	chunkChan := make(chan []byte)

	go func() {
		defer close(chunkChan)

		file, err := os.Open(filePath)
		if err != nil {
			return
		}
		defer file.Close()

		buffer := make([]byte, 64*1024) // 64KB chunks
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				chunkChan <- buffer[:n]
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
		}
	}()

	return fileInfo, chunkChan, nil
}

func (r *fileRepository) DeleteFile(fileID string) error {
	filePath := r.getFilePath(fileID)
	return os.Remove(filePath)
}

func generateFileID(recordID, filename string) string {
	hash := sha256.Sum256([]byte(recordID + filename))
	return hex.EncodeToString(hash[:])
}
