package httpServer

import (
	"context"
	"fmt"
	"net"

	"github.com/jxncyjq/stardust.mini/logs"
	"github.com/jxncyjq/stardust.mini/utils"
	"github.com/jxncyjq/stardust.mini/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GrpcServer gRPC服务器
type GrpcServer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	addr     string
	logger   *zap.Logger
	server   *grpc.Server
	listener net.Listener
}

// NewGrpcServer 创建gRPC服务器
func NewGrpcServer(configByte []byte, opts ...grpc.ServerOption) (*GrpcServer, error) {
	config, err := utils.Bytes2Struct[HttpServerConfig](configByte)
	if err != nil {
		return nil, err
	}

	uuid.InitWorker(config.WorkerID)

	ctx, cancel := context.WithCancel(context.Background())
	addr := fmt.Sprintf("%s:%d", config.Address, config.Port)

	return &GrpcServer{
		ctx:    ctx,
		cancel: cancel,
		addr:   addr,
		logger: logs.GetLogger("grpcServer"),
		server: grpc.NewServer(opts...),
	}, nil
}

// Server 获取grpc.Server用于注册服务
func (s *GrpcServer) Server() *grpc.Server {
	return s.server
}

// Startup 启动服务器
func (s *GrpcServer) Startup() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener

	s.logger.Info("gRPC server listened on:", zap.String("addr", s.addr))

	go func() {
		if err := s.server.Serve(listener); err != nil {
			s.logger.Error("gRPC server error:", zap.Error(err))
		}
	}()

	go func() {
		<-s.ctx.Done()
		s.Stop()
	}()

	return nil
}

// Stop 停止服务器
func (s *GrpcServer) Stop() {
	s.server.GracefulStop()
	s.logger.Info("gRPC server shutdown gracefully")
}

// Address 获取监听地址
func (s *GrpcServer) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}
