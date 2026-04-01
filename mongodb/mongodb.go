package mongodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// Config MongoDB 连接配置。
type Config struct {
	Name      string `json:"name"`       // 逻辑名称，多库时用于区分
	URI       string `json:"uri"`        // mongodb://user:pass@host:port/dbname?authSource=admin
	Database  string `json:"database"`   // 默认操作的数据库名
	MaxPool   uint64 `json:"max_pool"`   // 最大连接池大小
	MinPool   uint64 `json:"min_pool"`   // 最小连接池大小
	TimeoutS  int    `json:"timeout_s"`  // 单次操作超时（秒）
}

// SetDefaults 填充未设置的默认值。
func (c *Config) SetDefaults() {
	if c.MaxPool == 0 {
		c.MaxPool = 10
	}
	if c.MinPool == 0 {
		c.MinPool = 2
	}
	if c.TimeoutS == 0 {
		c.TimeoutS = 5
	}
}

// Validate 校验必填字段。
func (c *Config) Validate() error {
	if c.Name == "" {
		return errors.New("mongodb: config name is required")
	}
	if c.URI == "" {
		return errors.New("mongodb: uri is required")
	}
	if c.Database == "" {
		return errors.New("mongodb: database is required")
	}
	return nil
}

var (
	managerOnce   sync.Once
	manager       *MongoManager
	managerConfig []*Config
)

// Init 解析配置，支持单个对象或数组两种 JSON 格式，与 databases.Init 保持一致。
// 必须在 GetMongoManager() 之前调用。
func Init(config []byte) {
	if len(config) == 0 {
		panic("mongodb: config is empty")
	}

	var single Config
	if err := json.Unmarshal(config, &single); err == nil && single.Name != "" {
		managerConfig = []*Config{&single}
		return
	}

	var multiple []*Config
	if err := json.Unmarshal(config, &multiple); err != nil {
		panic(fmt.Errorf("mongodb: parse config failed: %w", err))
	}
	managerConfig = multiple
}

// GetMongoManager 返回全局单例 MongoManager，首次调用时初始化所有连接。
func GetMongoManager() *MongoManager {
	if len(managerConfig) == 0 {
		panic("mongodb: not initialized, call Init() first")
	}
	managerOnce.Do(func() {
		manager = newMongoManager(managerConfig)
	})
	return manager
}

// GetClient 快捷方法：获取指定名称的 MongoCli。
func GetClient(name string) MongoCli {
	return GetMongoManager().GetClient(name)
}
