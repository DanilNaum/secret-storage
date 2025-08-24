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

type GrpcAuthHandler struct {
	pb.UnimplementedAuthServiceServer
	usecases usecases
	grpcserver.Handler
}

func NewGrpcAuthHandler(usecases usecases) *GrpcAuthHandler {
	return &GrpcAuthHandler{usecases: usecases}
}
func (h *GrpcAuthHandler) RegisterServer(gRPC *grpc.Server) {
	pb.RegisterAuthServiceServer(gRPC, h)
}

func (h *GrpcAuthHandler) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
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
func (h *GrpcAuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{
		Success: true,
	}, nil
}
func (h *GrpcAuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
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
