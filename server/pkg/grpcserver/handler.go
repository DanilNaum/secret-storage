package grpcserver

import "google.golang.org/grpc"

type Handler struct {
}

func (h *Handler) RegisterServer(gRPC *grpc.Server) {
	
}
func (h *Handler) GetUnaryInterseptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{}
}

func (h *Handler) GetStreamInterseptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{}
}