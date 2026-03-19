// Package main 提供完整的微服务启动示例
// 展示如何使用 stardust.mini 框架的服务治理、弹性流控、RPC 和可观测性能力
package main

import (
	"time"

	httpServer "github.com/jxncyjq/stardust.mini/http_server"
	"github.com/jxncyjq/stardust.mini/http_server/middleware"
	"github.com/jxncyjq/stardust.mini/service"
)

// 示例：完整微服务启动流程
// 注意：实际运行需要配置文件和基础设施（Redis、etcd 等）
// 此文件仅用于展示框架使用方式和编译验证
func main() {
	// 1. 服务配置
	svcConf := service.ServiceConf{
		Name: "user-service",
		Mode: service.ModePro,
		Log:  service.LogConf{Level: "info"},
	}
	if err := svcConf.SetUp(); err != nil {
		panic(err)
	}

	// 2. 创建 RPC 服务（自动注入拦截器链：metric -> tracing -> breaker -> timeout）
	rpcServer, err := httpServer.NewGrpcServer(httpServer.GrpcServerConfig{
		ListenOn: "0.0.0.0:9090",
		Timeout:  5000,
	})
	if err != nil {
		panic(err)
	}
	// rpcServer.Server() 可用于注册 protobuf 服务
	// pb.RegisterUserServiceServer(rpcServer.Server(), &UserServiceImpl{})

	// 3. 中间件展示（实际使用时通过 httpServer.Use() 注入）
	_ = middleware.Metrics("user-service")
	_ = middleware.Tracing("user-service")
	_ = middleware.CircuitBreaker()
	_ = middleware.RateLimit(1000, 1000, "user-service:rl")
	_ = middleware.Timeout(3 * time.Second)
	_ = middleware.AdaptiveShedding()
	_ = middleware.MetricsHandler() // 用于 GET /metrics 端点

	// 4. 使用 ServiceGroup 统一管理生命周期
	sg := service.NewServiceGroup()
	sg.Add(rpcServer) // GrpcServer 直接实现 Service 接口
	// sg.Add(service.NewServerStarter(backend)) // HTTP 服务通过 ServerStarter 适配
	sg.Start() // 阻塞直到收到退出信号
}
