package zrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRpcServerConfig(t *testing.T) {
	conf := RpcServerConf{
		ListenOn: "0.0.0.0:9090",
		Timeout:  5000,
	}
	assert.Equal(t, "0.0.0.0:9090", conf.ListenOn)
}

func TestNewRpcServer(t *testing.T) {
	conf := RpcServerConf{
		ListenOn: "0.0.0.0:0",
		Timeout:  3000,
	}
	server, err := NewRpcServer(conf)
	assert.NoError(t, err)
	assert.NotNil(t, server)
}
