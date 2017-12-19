package manager

import (
	"github.com/Attsun1031/jobnetes/dao/db"
	"github.com/Attsun1031/jobnetes/utils/log"
)

type _ManagerConfig struct {
	DbConfig  *db.DbConfig
	LogConfig *log.LogConfig
}

var ManagerConfig *_ManagerConfig

func load() *_ManagerConfig {
	ManagerConfig = &_ManagerConfig{
		DbConfig:  db.LoadDbConfig(),
		LogConfig: log.LoadLogConfig(),
	}
	return ManagerConfig
}

func InitConfig() {
	ManagerConfig = load()
	log.SetupLogger(ManagerConfig.LogConfig)
	db.Connect(ManagerConfig.DbConfig)
}