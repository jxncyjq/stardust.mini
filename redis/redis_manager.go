package redis

import (
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

type RedisManager struct {
	mu        sync.Mutex
	redisCmds map[string]RedisCmd
}

var (
	manager       *RedisManager
	managerOnce   sync.Once
	managerConfig []*Config
)

func Init(config []byte) {
	err := json.Unmarshal(config, &managerConfig)
	if err != nil {
		panic(err)
	}
}

func GetRedisManager() *RedisManager {
	managerOnce.Do(func() {
		manager = &RedisManager{
			redisCmds: make(map[string]RedisCmd),
		}
		for _, cfg := range managerConfig {
			cli, err := NewRedisCmd(cfg)
			if err != nil {
				panic(err)
			}
			manager.redisCmds[cfg.Name] = cli
		}
	})
	return manager
}

func (m *RedisManager) GetRedisCmd(name string) RedisCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cmd, ok := m.redisCmds[name]; ok {
		return cmd
	}
	return nil
}

func (m *RedisManager) GetRedisView(name, prefix string, logger *zap.Logger) RedisCli {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cmd, ok := m.redisCmds[name]; ok {
		return NewRedisView(cmd, prefix, logger)
	}
	return nil
}
