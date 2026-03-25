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

// ensureRedisInitialized creates all configured Redis clients once.
//
// The first configured client becomes the package default returned by
// GetRedisDb, and every configured client is also registered in the manager for
// named lookup. When no config has been loaded yet, initialization is skipped.
func ensureRedisInitialized() {
	if len(managerConfig) == 0 {
		return
	}

	managerOnce.Do(func() {
		redisCmds := make(map[string]RedisCmd, len(managerConfig))
		for index, cfg := range managerConfig {
			cli, err := NewRedisCmd(cfg)
			if err != nil {
				panic(err)
			}
			if index == 0 {
				redisCon = cli
			}
			redisCmds[cfg.Name] = cli
		}
		manager = &RedisManager{
			redisCmds: redisCmds,
		}
	})
}

// GetRedisManager returns the named Redis client manager.
//
// If Init has not loaded any config yet, GetRedisManager returns an empty
// manager so callers can safely retry after configuration is available.
func GetRedisManager() *RedisManager {
	ensureRedisInitialized()
	if manager == nil {
		return &RedisManager{redisCmds: make(map[string]RedisCmd)}
	}
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
