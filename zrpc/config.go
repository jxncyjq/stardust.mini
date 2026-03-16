package zrpc

import (
	"github.com/jxncyjq/stardust.mini/register"
)

// RpcServerConf gRPC 服务器配置
type RpcServerConf struct {
	ListenOn string               `json:"listen_on" toml:"listen_on"` // 监听地址
	Timeout  int64                `json:"timeout" toml:"timeout"`     // 超时(ms)
	Etcd     *register.EtcdConfig `json:"etcd,omitempty" toml:"etcd,omitempty"`
}

// RpcClientConf gRPC 客户端配置
type RpcClientConf struct {
	Target  string               `json:"target" toml:"target"`   // 直连地址（与 Etcd 二选一）
	Timeout int64                `json:"timeout" toml:"timeout"` // 超时(ms)
	Etcd    *register.EtcdConfig `json:"etcd,omitempty" toml:"etcd,omitempty"`
}
