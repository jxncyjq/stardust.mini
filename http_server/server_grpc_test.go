package httpServer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrpcServerConfig(t *testing.T) {
	conf := GrpcServerConfig{
		ListenOn: "0.0.0.0:9090",
		Timeout:  5000,
	}
	assert.Equal(t, "0.0.0.0:9090", conf.ListenOn)
	assert.Equal(t, int64(5000), conf.Timeout)
}

func TestNewGrpcServer(t *testing.T) {
	conf := GrpcServerConfig{
		ListenOn: "0.0.0.0:0",
		Timeout:  3000,
	}
	server, err := NewGrpcServer(conf)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.NotNil(t, server.Server())
}

func TestNewGrpcServer_DefaultTimeout(t *testing.T) {
	conf := GrpcServerConfig{
		ListenOn: "0.0.0.0:0",
		Timeout:  0,
	}
	server, err := NewGrpcServer(conf)
	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestGrpcServer_StartupAndStop(t *testing.T) {
	conf := GrpcServerConfig{
		ListenOn: "127.0.0.1:0",
		Timeout:  3000,
	}
	server, err := NewGrpcServer(conf)
	assert.NoError(t, err)

	err = server.Startup()
	assert.NoError(t, err)
	assert.NotEmpty(t, server.Address())

	server.Stop()
}

func TestGrpcClient_DirectConnect(t *testing.T) {
	conf := GrpcClientConfig{
		Target:  "localhost:9999",
		Timeout: 3000,
	}
	client, err := NewGrpcClient(conf)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Conn())
	client.Close()
}

func TestParseAddr(t *testing.T) {
	host, port := parseAddr("127.0.0.1:9090")
	assert.Equal(t, "127.0.0.1", host)
	assert.Equal(t, 9090, port)
}
