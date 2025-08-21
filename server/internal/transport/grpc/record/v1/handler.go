package record

import (
	"context"
	"errors"
	"io"

	// "fmt"
	// "go/token"
	// "strings"

	"github.com/DanilNaum/secret-storage-server/internal/entity"
	"github.com/DanilNaum/secret-storage-server/pkg/grpcserver"
	pb "github.com/DanilNaum/secret-storage-server/pkg/proto"
	"google.golang.org/grpc"

	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
	"github.com/DanilNaum/secret-storage-server/internal/filerepository"
)

type JWTManager[T any] interface {
	ParseToken(tokenString string) (T, error)
}
type usecase interface {
	CreateRecord(context.Context, int, *entity.Record) (string, error)
	DeleteRecord(context.Context, int, string) error
	GetRecord(context.Context, int, string) (*entity.Record, error)
	ListRecords(context.Context, int) ([]*entity.RecordInfo, error)
	UpdateRecord(context.Context, int, *entity.Record) error
}
type fileRepo interface {
	SaveFile(metadata *filerepository.FileMetadata, chunks <-chan []byte, errChan <-chan error) error
	GetFile(fileID string) (*filerepository.FileInfo, <-chan []byte, error)
}

type grpcRecordHandler struct {
	pb.UnimplementedRecordServiceServer
	jwtManager JWTManager[int]
	usecase    usecase
	grpcserver.Handler
	fileRepo fileRepo
}

func NewGrpcRecordHandler(jwtManager JWTManager[int], usecase usecase, fileRepo fileRepo) *grpcRecordHandler {
	return &grpcRecordHandler{jwtManager: jwtManager, usecase: usecase, fileRepo: fileRepo}
}

type request interface {
	GetAuthToken() string
}

func (h *grpcRecordHandler) GetUnaryInterseptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{h.authUnaryInterceptor}
}

func (h *grpcRecordHandler) GetStreamInterseptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{h.authStreamInterceptor}
}

func (h *grpcRecordHandler) RegisterServer(gRPC *grpc.Server) {
	pb.RegisterRecordServiceServer(gRPC, h)
}

func (h *grpcRecordHandler) CreateRecord(ctx context.Context, req *pb.CreateRecordRequest) (*pb.CreateRecordResponse, error) {
	userId, err := h.getUserId(ctx)
	if err != nil {
		return nil, err
	}
	id, err := h.usecase.CreateRecord(ctx, userId, recordToEntityRecord(req.Record))
	if err != nil {
		return &pb.CreateRecordResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.CreateRecordResponse{
		Success:  true,
		ServerId: id,
	}, nil

}

func (h *grpcRecordHandler) DeleteRecord(ctx context.Context, req *pb.DeleteRecordRequest) (*pb.DeleteRecordResponse, error) {
	userId, err := h.getUserId(ctx)
	if err != nil {
		return nil, err
	}
	err = h.usecase.DeleteRecord(ctx, userId, req.ServerId)
	if err != nil {
		return &pb.DeleteRecordResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.DeleteRecordResponse{
		Success: true,
	}, nil
}

func (h *grpcRecordHandler) DownloadFile(req *pb.FileDownloadRequest, stream grpc.ServerStreamingServer[pb.FileDownloadResponse]) error {

	userId, err := h.getUserId(stream.Context())
	if err != nil {
		return err
	}
	_, err = h.usecase.GetRecord(stream.Context(), userId, req.RecordId)
	if err != nil {
		return err
	}

	info, chunks, err := h.fileRepo.GetFile(req.RecordId)
	if err != nil {
		return err
	}
	err = stream.Send(&pb.FileDownloadResponse{
		Data: &pb.FileDownloadResponse_Info{
			Info: &pb.FileMetadata{
				RecordId:    req.RecordId,
				Filename:    info.Filename,
				FileSize:    info.FileSize,
				ContentType: info.ContentType,
			},
		},
	})
	if err != nil {
		return err
	}

	for chunk := range chunks {
		err := stream.Send(&pb.FileDownloadResponse{
			Data: &pb.FileDownloadResponse_Chunk{
				Chunk: chunk,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *grpcRecordHandler) GetRecord(ctx context.Context, req *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	userId, err := h.getUserId(ctx)
	if err != nil {
		return nil, err
	}
	record, err := h.usecase.GetRecord(ctx, userId, req.ServerId)
	if err != nil {
		return &pb.GetRecordResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.GetRecordResponse{
		Success: true,
		Record:  recordFromEntityRecord(record),
	}, nil
}

func (h *grpcRecordHandler) ListRecords(ctx context.Context, req *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
	userId, err := h.getUserId(ctx)
	if err != nil {
		return nil, err
	}
	records, err := h.usecase.ListRecords(ctx, userId)
	if err != nil {
		return &pb.ListRecordsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.ListRecordsResponse{
		Success: true,
		Records: serverRecordsFromEntityRecordsInfo(records),
	}, nil
}

func (h *grpcRecordHandler) UpdateRecord(ctx context.Context, req *pb.UpdateRecordRequest) (*pb.UpdateRecordResponse, error) {
	userId, err := h.getUserId(ctx)
	if err != nil {
		return nil, err
	}
	err = h.usecase.UpdateRecord(ctx, userId, recordToEntityRecord(req.Record))
	if err != nil {
		return &pb.UpdateRecordResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.UpdateRecordResponse{
		Success: true,
	}, nil
}

func (h *grpcRecordHandler) UploadFile(stream grpc.ClientStreamingServer[pb.FileUploadRequest, pb.FileUploadResponse]) error {
	userId, err := h.getUserId(stream.Context())
	if err != nil {
		return err
	}

	req, err := stream.Recv()
	if err != nil {
		return err
	}
	metadata, ok := req.Data.(*pb.FileUploadRequest_Metadata)
	if !ok {
		return errors.New("expect metadata, got something else")
	}
	// err = stream.SendMsg(&pb.FileUploadResponse{
	// 	Success: true,
	// })

	_, err = h.usecase.GetRecord(stream.Context(), userId, metadata.Metadata.RecordId)
	if err != nil {
		return err
	}

	chunc := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
			defer close(chunc)
			defer close(errChan)
	LOOP:
		for {
		
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break LOOP
			}
			if err != nil {
				stream.SendMsg(&pb.FileUploadResponse{
					Success: false,
					Message: err.Error(),
				},
				)
				errChan <- err
				return 
			}

			switch r := req.Data.(type) {
			case *pb.FileUploadRequest_Chunk:
				chunc <- r.Chunk
			case *pb.FileUploadRequest_Metadata:
				errChan <- errors.New("expect chunk, got metadata")
				return
			}

		}
	}()
	err = h.fileRepo.SaveFile(
		&filerepository.FileMetadata{RecordID: metadata.Metadata.GetRecordId(),
			FileInfo: &filerepository.FileInfo{
				Filename:    metadata.Metadata.Filename,
				FileSize:    metadata.Metadata.FileSize,
				ContentType: metadata.Metadata.ContentType,
			},
		}, chunc, errChan)
	if err != nil {
		stream.SendMsg(&pb.FileUploadResponse{
			Success: false,
			Message: err.Error(),
		},
		)
		return err
	}

	stream.SendMsg(&pb.FileUploadResponse{
		Success: true,
	})
	return nil
}
