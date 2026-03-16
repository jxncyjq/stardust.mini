package zrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRpcClient_DirectConnect(t *testing.T) {
	conf := RpcClientConf{
		Target:  "localhost:9999",
		Timeout: 3000,
	}
	client, err := NewRpcClient(conf)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Conn())
	client.Close()
}
