package record

import (
	"context"
	// "fmt"
	// "go/token"
	// "strings"

	pb "github.com/DanilNaum/secret-storage-server/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// authUnaryInterceptor - миделвара для аутентификации пользователей через JWT токен
func (h *GrpcRecordHandler) authUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Пропускаем аутентификацию для методов других пакетов
	if info.FullMethod != pb.RecordService_CreateRecord_FullMethodName &&
		info.FullMethod != pb.RecordService_GetRecord_FullMethodName &&
		info.FullMethod != pb.RecordService_UpdateRecord_FullMethodName &&
		info.FullMethod != pb.RecordService_DeleteRecord_FullMethodName &&
		info.FullMethod != pb.RecordService_ListRecords_FullMethodName {
		return handler(ctx, req)
	}

	request, ok := req.(request)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "request doesn't contains token")
	}
	token := request.GetAuthToken()
	userId, err := h.jwtManager.ParseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Добавляем ID пользователя в контекст
	newCtx := context.WithValue(ctx, "user_id", userId)
	return handler(newCtx, req)

}

type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStreamWrapper) Context() context.Context {
	return s.ctx
}
func (h *GrpcRecordHandler) authStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if info.FullMethod != pb.RecordService_UploadFile_FullMethodName &&
		info.FullMethod != pb.RecordService_DownloadFile_FullMethodName {
		return handler(srv, ss)
	}

	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	if len(md["authorization"]) == 0 {
		return status.Error(codes.Unauthenticated, "authorization token is required")
	}
	token := md["authorization"][0]
	userId, err := h.jwtManager.ParseToken(token)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	newCtx := context.WithValue(ctx, "user_id", userId)
	return handler(srv, &serverStreamWrapper{ServerStream: ss,
		ctx: newCtx})
}

// getUserId извлекает ID пользователя из контекста
func (h *GrpcRecordHandler) getUserId(ctx context.Context) (int, error) {
	userId, ok := ctx.Value("user_id").(int)
	if !ok {
		return 0, status.Errorf(codes.Internal, "failed to get user id from context")
	}
	return userId, nil
}
