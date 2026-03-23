package databases

import (
	"encoding/json"
	"fmt"
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
	if len(config) == 0 {
		panic("database config is empty")
	}

	var single Config
	if err := json.Unmarshal(config, &single); err == nil && single.Name != "" {
		managerConfig = []*Config{&single}
		return
	}

	var multiple []*Config
	if err := json.Unmarshal(config, &multiple); err != nil {
		panic(fmt.Errorf("parse database config failed: %w", err))
	}
	managerConfig = multiple
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
			cfg.SetDefaults()
			if err := cfg.Validate(); err != nil {
				panic("Invalid database config: " + err.Error())
			}
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
