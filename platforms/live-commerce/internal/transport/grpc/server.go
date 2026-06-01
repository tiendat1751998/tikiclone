package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/tikiclone/tiki/platforms/live-commerce/internal/application"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	service    *application.LiveCommerceService
	logger     *zap.Logger
}

func NewServer(svc *application.LiveCommerceService) *Server {
	s := &Server{
		service: svc,
		logger:  zap.L(),
	}
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.MaxRecvMsgSize(4 * 1024 * 1024),
		grpc.MaxSendMsgSize(4 * 1024 * 1024),
	)
	grpc_health_v1.RegisterHealthServer(s.grpcServer, &healthServer{})
	reflection.Register(s.grpcServer)
	return s
}

func (s *Server) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}
	s.logger.Info("starting grpc server", zap.Int("port", port))
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

func (s *Server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)
	s.logger.Debug("grpc", zap.String("method", info.FullMethod), zap.Duration("duration", duration), zap.Error(err))
	return resp, err
}

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING})
}
