package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pb "github.com/DanilNaum/secret-storage-client/pkg/proto"
	"google.golang.org/grpc/metadata"
)

const (
	// ChunkSize defines the size of file chunks for streaming (1MB)
	ChunkSize = 1024 * 1024
)

// UploadFile uploads a file to the server using streaming
func (c *Client) UploadFile(recordID, filePath string) error {
	if c.authToken == "" {
		return fmt.Errorf("not authenticated")
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	md := metadata.New(map[string]string{
		"authorization": c.authToken,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := c.createContext(ctx)
	defer cancel()

	// Create upload stream
	stream, err := c.recordClient.UploadFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload stream: %w", err)
	}

	// Send metadata first
	metadata := &pb.FileUploadRequest{
		Data: &pb.FileUploadRequest_Metadata{
			Metadata: &pb.FileMetadata{
				RecordId:    recordID,
				Filename:    filepath.Base(filePath),
				FileSize:    fileInfo.Size(),
				ContentType: detectContentType(filePath),
			},
		},
	}

	if err := stream.Send(metadata); err != nil {
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	// Send file chunks
	buffer := make([]byte, ChunkSize)
	LOOP:for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break LOOP
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		chunk := &pb.FileUploadRequest{
			Data: &pb.FileUploadRequest_Chunk{
				Chunk: buffer[:n],
			},
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close stream: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("upload failed: %s", resp.Message)
	}

	return nil
}

// DownloadFile downloads a file from the server using streaming
func (c *Client) DownloadFile(recordID, savePath string) (chan int, chan error) {

	progres := make(chan int, 1)
	errChan := make(chan error, 1)

	go func() {
		fileSize := -1.
		progres <- 0
		defer close(progres)
		defer close(errChan)

		if c.authToken == "" {
			errChan <- fmt.Errorf("not authenticated")
			return
		}
		md := metadata.New(map[string]string{
			"authorization": c.authToken,
		})
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		req := &pb.FileDownloadRequest{
			AuthToken: c.authToken,
			RecordId:  recordID,
		}

		stream, err := c.recordClient.DownloadFile(ctx, req)
		if err != nil {
			errChan <- fmt.Errorf("failed to create download stream: %w", err)
			return
		}

		// Create output file
		outFile, err := os.Create(savePath + "/" + recordID)
		if err != nil {
			errChan <- fmt.Errorf("failed to create output file: %w", err)
			return
		}

		var fileInfo *pb.FileMetadata
		var totalSize int64
	LOOP:
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {

				break LOOP
			}
			if err != nil {
				errChan <- fmt.Errorf("failed to receive data: %w", err)
				outFile.Close()
				return
			}

			switch data := resp.Data.(type) {
			case *pb.FileDownloadResponse_Info:
				fileInfo = data.Info
				fileSize = float64(fileInfo.FileSize)
				// You could validate file info here if needed
			case *pb.FileDownloadResponse_Chunk:
				chunk := data.Chunk
				totalSize += int64(len(chunk))
				if fileSize > 0 {
					progres <- int(float64(totalSize) / fileSize * 100)
				}
				if _, err := outFile.Write(chunk); err != nil {
					errChan <- fmt.Errorf("failed to write chunk: %w", err)
					return
				}
			}
		}
		outFile.Close()
		if fileInfo == nil {
			errChan <- fmt.Errorf("no file info received")
			return
		}
		if fileInfo.FileSize != totalSize {
			errChan <- fmt.Errorf("file size mismatch: expected %d, got %d", fileInfo.FileSize, totalSize)
			return
		}

		err = os.Rename(savePath+"/"+recordID, savePath+"/"+fileInfo.Filename)
		if err != nil {
			errChan <- fmt.Errorf("failed to rename file: %w", err)
			return
		}
	}()

	return progres, errChan
}

// detectContentType detects the content type of a file based on its extension
func detectContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".txt":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".zip":
		return "application/zip"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}
