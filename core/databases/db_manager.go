package databases

import (
	"encoding/json"
	"sync"
)

type DatabaseManager struct {
	mu sync.RWMutex
	db map[string]DBInterface
}

var (
	manager       *DatabaseManager
	managerOnce   sync.Once
	managerConfig []*Config
)

func Init(config []byte) {
	err := json.Unmarshal(config, &managerConfig)
	if err != nil {
		panic(err)
	}
}

func GetDatabaseManager() *DatabaseManager {
	if managerConfig == nil || len(managerConfig) == 0 {
		panic("DatabaseManager not initialized. Call Init(config) first.")
	}
	managerOnce.Do(func() {
		manager = &DatabaseManager{
			db: make(map[string]DBInterface),
		}
		for _, cfg := range managerConfig {
			db := NewDBInterface(cfg)
			manager.db[cfg.Name] = db
		}
	})
	return manager
}

func (m *DatabaseManager) GetDBDao(name string) BaseDao {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if db, exists := m.db[name]; exists {
		return NewBaseDao(db)
	}
	return nil
}

func (m *DatabaseManager) GetDbInterface(name string) DBInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if db, exists := m.db[name]; exists {
		return db
	}
	return nil
}
