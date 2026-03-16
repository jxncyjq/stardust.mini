package zrpc

import (
	"net"
	"time"

	"github.com/jxncyjq/stardust.mini/register"
	"github.com/jxncyjq/stardust.mini/zrpc/interceptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RpcServer 增强的 gRPC 服务器（参照 go-zero zrpc.RpcServer）
type RpcServer struct {
	conf     RpcServerConf
	server   *grpc.Server
	listener net.Listener
	registry *register.ServiceRegistry
	logger   *zap.Logger
}

// NewRpcServer 创建 RPC 服务器，自动注入拦截器链
func NewRpcServer(conf RpcServerConf, opts ...grpc.ServerOption) (*RpcServer, error) {
	timeout := time.Duration(conf.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	// 内置拦截器链（顺序很重要）
	chainedInterceptors := grpc.ChainUnaryInterceptor(
		interceptor.UnaryMetricInterceptor(),
		interceptor.UnaryTracingInterceptor("rpc"),
		interceptor.UnaryBreakerInterceptor(),
		interceptor.UnaryTimeoutInterceptor(timeout),
	)
	opts = append([]grpc.ServerOption{chainedInterceptors}, opts...)

	return &RpcServer{
		conf:   conf,
		server: grpc.NewServer(opts...),
		logger: getLoggerSafe("rpc_server"),
	}, nil
}

// Server 获取原生 grpc.Server 用于注册服务
func (s *RpcServer) Server() *grpc.Server {
	return s.server
}

// Start 实现 Service 接口
func (s *RpcServer) Start() {
	if err := s.Startup(); err != nil {
		if s.logger != nil {
			s.logger.Fatal("rpc server start failed", zap.Error(err))
		}
	}
}

// Startup 启动服务器
func (s *RpcServer) Startup() error {
	listener, err := net.Listen("tcp", s.conf.ListenOn)
	if err != nil {
		return err
	}
	s.listener = listener
	if s.logger != nil {
		s.logger.Info("rpc server listening", zap.String("addr", listener.Addr().String()))
	}

	// 服务注册
	if s.conf.Etcd != nil {
		reg, err := register.NewEtcdRegister(s.conf.Etcd)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("etcd register failed", zap.Error(err))
			}
		} else {
			s.registry = register.NewServiceRegistry(reg)
			host, port := parseAddr(listener.Addr().String())
			s.registry.Register(s.conf.Etcd.ServiceName, host, port, s.conf.Etcd.Tags, nil)
		}
	}

	go func() {
		if err := s.server.Serve(listener); err != nil {
			if s.logger != nil {
				s.logger.Error("rpc server error", zap.Error(err))
			}
		}
	}()
	return nil
}

// Stop 优雅关闭
func (s *RpcServer) Stop() {
	if s.registry != nil {
		s.registry.Close()
	}
	s.server.GracefulStop()
	if s.logger != nil {
		s.logger.Info("rpc server stopped")
	}
}

// Address 获取监听地址
func (s *RpcServer) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.conf.ListenOn
}

func parseAddr(addr string) (string, int) {
	host, portStr, _ := net.SplitHostPort(addr)
	port := 0
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}
	return host, port
}
