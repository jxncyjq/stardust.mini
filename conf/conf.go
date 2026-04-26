package conf

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	config   map[string]interface{}
	vp       *viper.Viper
	appName  string
	appVer   string
	redisKey string
	mu       sync.RWMutex
	initOnce sync.Once
	initErr  error
)

func resolveConfigPath() string {
	runConfig := os.Getenv("runConfig")
	if runConfig != "" {
		return runConfig
	}

	isDebug := os.Getenv("ISDEBUG")
	if isDebug != "" && isDebug != "0" {
		if devConf := os.Getenv("devConf"); devConf != "" {
			return devConf
		}
	}

	return os.Getenv("prodConf")
}

func Init() {
	initOnce.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		configPath := resolveConfigPath()
		if configPath == "" {
			initErr = errors.New("config path is empty, set runConfig or devConf/prodConf")
			return
		}

		v := viper.New()
		v.SetConfigFile(configPath)
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		err := v.ReadInConfig()
		if err != nil {
			initErr = err
			return
		}

		globalInfo := v.GetStringMap("global")
		if len(globalInfo) == 0 {
			initErr = errors.New("global configuration is missing in the config file")
			return
		}

		if !v.IsSet("global.app_name") {
			initErr = errors.New("app name is missing in the global configuration")
			return
		}
		appNameValue := v.GetString("global.app_name")

		if !v.IsSet("global.app_version") {
			initErr = errors.New("app version is missing in the global configuration")
			return
		}
		appVersionValue := v.GetString("global.app_version")

		if !v.IsSet("global.redis_key_prefix") {
			initErr = errors.New("redis key prefix is missing in the global configuration")
			return
		}
		redisKeyPrefixValue := v.GetString("global.redis_key_prefix")

		vp = v
		config = v.AllSettings()
		appName = appNameValue
		appVer = appVersionValue
		redisKey = redisKeyPrefixValue
	})

	if initErr != nil {
		panic(initErr)
	}
}

func GetAppName() string {
	mu.RLock()
	name := appName
	mu.RUnlock()
	if name != "" {
		return name
	}
	Init()
	mu.RLock()
	defer mu.RUnlock()
	return appName
}

func GetAppVersion() string {
	mu.RLock()
	version := appVer
	mu.RUnlock()
	if version != "" {
		return version
	}
	Init()
	mu.RLock()
	defer mu.RUnlock()
	return appVer
}

func GetRedisKeyPrefix() string {
	mu.RLock()
	prefix := redisKey
	mu.RUnlock()
	if prefix != "" {
		return prefix
	}
	Init()
	mu.RLock()
	defer mu.RUnlock()
	return redisKey
}

func Get(key string) []byte {
	Init()

	mu.RLock()
	defer mu.RUnlock()

	if value, exists := config[key]; exists {
		if strValue, ok := value.(string); ok {
			return []byte(strValue)
		}
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		return bytes
	}

	if vp != nil && vp.IsSet(key) {
		value := vp.Get(key)
		if strValue, ok := value.(string); ok {
			return []byte(strValue)
		}
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		return bytes
	}

	return nil
}

func GetConfigInstance() *viper.Viper {
	mu.RLock()
	v := vp
	mu.RUnlock()
	if v == nil {
		Init()
		mu.RLock()
		v = vp
		mu.RUnlock()
	}
	return v
}
