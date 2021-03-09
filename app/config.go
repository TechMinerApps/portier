package app

import (
	"github.com/TechMinerApps/portier/modules/bot"
	"github.com/TechMinerApps/portier/modules/database"
)

// Config is the configuration used in viper
type Config struct {
	DB       database.DBConfig
	Telegram bot.Config
	Template string
	Log      logConfig
}

type logConfig struct {
	Mode string
	Path string
}
