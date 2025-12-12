package httpServer

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestGrpcServerCreate(t *testing.T) {
	config := []byte(`{"address":"127.0.0.1","port":0,"mode":"grpc"}`)
	srv, err := NewGrpcServer(config)
	if err != nil {
		t.Fatalf("Failed to create GrpcServer: %v", err)
	}

	if srv.Server() == nil {
		t.Error("Expected grpc.Server not nil")
	}
}

func TestGrpcServerStartup(t *testing.T) {
	config := []byte(`{"address":"127.0.0.1","port":0,"mode":"grpc"}`)
	srv, err := NewGrpcServer(config)
	if err != nil {
		t.Fatalf("Failed to create GrpcServer: %v", err)
	}

	// 注册健康检查服务
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv.Server(), healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	err = srv.Startup()
	if err != nil {
		t.Fatalf("Startup failed: %v", err)
	}
	defer srv.Stop()

	time.Sleep(100 * time.Millisecond)

	addr := srv.Address()
	if addr == "" {
		t.Error("Expected address not empty")
	}
	t.Logf("gRPC server listening on: %s", addr)

	// 测试连接
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 测试健康检查
	client := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("Expected SERVING status, got %v", resp.Status)
	}
	t.Log("gRPC health check passed")
}

func TestServerFactory(t *testing.T) {
	// 测试 gin 模式
	ginConfig := []byte(`{"address":"127.0.0.1","port":0,"mode":"gin"}`)
	ginSrv, err := NewServer(ginConfig)
	if err != nil {
		t.Fatalf("Failed to create gin server: %v", err)
	}
	if _, ok := ginSrv.(*HttpServer); !ok {
		t.Error("Expected HttpServer for gin mode")
	}

	// 测试 grpc 模式
	grpcConfig := []byte(`{"address":"127.0.0.1","port":0,"mode":"grpc"}`)
	grpcSrv, err := NewServer(grpcConfig)
	if err != nil {
		t.Fatalf("Failed to create grpc server: %v", err)
	}
	if _, ok := grpcSrv.(*GrpcServer); !ok {
		t.Error("Expected GrpcServer for grpc mode")
	}

	// 测试默认模式 (gin)
	defaultConfig := []byte(`{"address":"127.0.0.1","port":0}`)
	defaultSrv, err := NewServer(defaultConfig)
	if err != nil {
		t.Fatalf("Failed to create default server: %v", err)
	}
	if _, ok := defaultSrv.(*HttpServer); !ok {
		t.Error("Expected HttpServer for default mode")
	}
}

func TestServerModeConfig(t *testing.T) {
	if ModeGin != "gin" {
		t.Errorf("Expected ModeGin = 'gin', got '%s'", ModeGin)
	}
	if ModeGrpc != "grpc" {
		t.Errorf("Expected ModeGrpc = 'grpc', got '%s'", ModeGrpc)
	}
}
