package conf

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type AppCfg struct {
	Database struct {
		ShowSql        bool     `json:"show_sql" toml:"show_sql" yaml:"show_sql"`
		MaxIdle        int      `json:"max_idle" toml:"max_idle" yaml:"max_idle"`
		MaxConn        int      `json:"max_conn" toml:"max_conn" yaml:"max_conn"`
		Master         string   `json:"master" toml:"master" yaml:"master"`
		Slaves         []string `json:"slaves" toml:"slaves" yaml:"slaves"`
		UseMasterSlave bool     `json:"use_master_slave" toml:"use_master_slave" yaml:"use_master_slave"`
		DbType         string   `json:"db_type" toml:"db_type" yaml:"db_type"`
	} `json:"database" toml:"database" yaml:"database"`
}

func InitCfg(fn string) (*AppCfg, error) {
	app := &AppCfg{}
	ext := filepath.Ext(fn)

	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	switch ext {
	case ".json":
		err = json.Unmarshal(data, app)
	case ".toml":
		_, err = toml.Decode(string(data), app)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, app)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	color.Red("Configuration file loaded successfully")
	color.Red(fmt.Sprintf("config %v", app))
	return app, nil
}
