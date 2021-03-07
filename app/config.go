package app

import (
	"github.com/TechMinerApps/portier/modules/bot"
	"github.com/TechMinerApps/portier/modules/database"
	"github.com/TechMinerApps/portier/modules/log"
	"github.com/TechMinerApps/portier/modules/render"
)

// Config is the configuration used in viper
type Config struct {
	DB       database.DBConfig
	Telegram bot.Config
	Template render.Config
	Log      logConfig
}

type logConfig struct {
	Mode log.LoggerType
	Path string
}
