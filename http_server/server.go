package httpServer

import (
	"github.com/jxncyjq/stardust.mini/utils"
)

// Server 服务器接口
type Server interface {
	Startup() error
	Stop()
}

// NewServer 根据配置创建服务器
func NewServer(configByte []byte) (Server, error) {
	config, err := utils.Bytes2Struct[HttpServerConfig](configByte)
	if err != nil {
		return nil, err
	}

	switch config.Mode {
	case ModeGrpc:
		return NewGrpcServer(configByte)
	default:
		return NewHttpServer(configByte)
	}
}
