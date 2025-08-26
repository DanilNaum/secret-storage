package grpcserver

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type handler interface {
	RegisterServer(gRPC *grpc.Server)
	GetUnaryInterseptors() []grpc.UnaryServerInterceptor
	GetStreamInterseptors() []grpc.StreamServerInterceptor
}

type GrpcServer struct {
	port   int
	server *grpc.Server
	notify chan error
}

func NewGrpcServer(port int, handlers ...handler) *GrpcServer {
	unaryInterseptors := make([]grpc.UnaryServerInterceptor, 0, 4)
	for _, handler := range handlers {
		unaryInterseptors = append(unaryInterseptors, handler.GetUnaryInterseptors()...)
	}
	streamInterseptors := make([]grpc.StreamServerInterceptor, 0, 4)
	for _, handler := range handlers {
		streamInterseptors = append(streamInterseptors, handler.GetStreamInterseptors()...)
	}
	grpcServ := grpc.NewServer(grpc.ChainUnaryInterceptor(unaryInterseptors...), grpc.ChainStreamInterceptor(streamInterseptors...))
	reflection.Register(grpcServ)

	for _, handler := range handlers {
		handler.RegisterServer(grpcServ)
	}

	return &GrpcServer{
		port:   port,
		server: grpcServ,
		notify: make(chan error),
	}
}

func (s *GrpcServer) Run(ctx context.Context) {
	list, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		s.notify <- err
		return
	}
	go func() {
		s.notify <- s.server.Serve(list)
		close(s.notify)
	}()
	go func() {
		<-ctx.Done()
		s.server.GracefulStop()
		s.notify <- ctx.Err()
	}()

}

func (s *GrpcServer) Notify() <-chan error {
	return s.notify
}
