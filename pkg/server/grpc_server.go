package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (s *DualServer) initGRPCServer() error {
	interceptors := s.middlewareRegistry.GetGRPCInterceptors()
	interceptors = append(interceptors, s.unaryInterceptor)

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.chainGRPCInterceptors(interceptors)),
	)

	reflection.Register(s.grpcServer)

	for _, service := range s.services {
		if err := s.registerGRPCService(service); err != nil {
			return fmt.Errorf("failed to register gRPC service %T: %w", service, err)
		}
	}

	return nil
}

func (s *DualServer) chainGRPCInterceptors(interceptors []grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var err error
		var resp any

		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			currentInterceptor := interceptors[i]
			nextChain := chain
			chain = func(ctx context.Context, req any) (any, error) {
				resp, err = currentInterceptor(ctx, req, info, nextChain)
				return resp, err
			}
		}

		return chain(ctx, req)
	}
}

func (s *DualServer) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	s.logger.Info("gRPC call",
		"method", info.FullMethod,
		"request", fmt.Sprintf("%+v", req),
	)

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	s.logger.Info("gRPC call completed",
		"method", info.FullMethod,
		"duration", duration,
		"error", err,
	)

	return resp, err
}
