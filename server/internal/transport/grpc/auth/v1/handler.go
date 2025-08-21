package auth

import (
	"context"

	"github.com/DanilNaum/secret-storage-server/internal/entity"
	"github.com/DanilNaum/secret-storage-server/pkg/grpcserver"
	pb "github.com/DanilNaum/secret-storage-server/pkg/proto"
	"google.golang.org/grpc"
)

type usecases interface {
	Authenticate(ctx context.Context, login, password string) (*entity.AuthData, error)
	Register(ctx context.Context, login, password string) (*entity.AuthData, error)
}

type grpcAuthHandler struct {
	pb.UnimplementedAuthServiceServer
	usecases usecases
	grpcserver.Handler
}

func NewGrpcAuthHandler(usecases usecases) *grpcAuthHandler {
	return &grpcAuthHandler{usecases: usecases}
}
func (h *grpcAuthHandler) RegisterServer(gRPC *grpc.Server) {
	pb.RegisterAuthServiceServer(gRPC, h)
}

func (h *grpcAuthHandler) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	authData, err := h.usecases.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.AuthenticateResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.AuthenticateResponse{
		Success:   true,
		AuthToken: authData.JWTToken,
		Salt:      authData.Salt,
	}, nil
}

// I deside to use jwt token for authentication, so it is imposible to logout
func (h *grpcAuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{
		Success: true,
	}, nil
}
func (h *grpcAuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	authData, err := h.usecases.Register(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.RegisterResponse{
		Success:   true,
		AuthToken: authData.JWTToken,
		Salt:      authData.Salt,
	}, nil
}
