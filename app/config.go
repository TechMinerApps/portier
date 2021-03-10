package app

import (
	"github.com/TechMinerApps/portier/modules/bot"
)

// DefaultConfig is the default config of Portier
var DefaultConfig Config = Config{
	DB:       dbConfig{Type: "sqlite", Path: "portier.db", Username: "portier", Password: "portier", Host: "localhost", Port: 3306, DBName: "portier"},
	Telegram: bot.Config{Token: ""},
	Template: "",
	Log:      logConfig{Mode: "", Path: ""},
	BuntDB:   buntDBConfig{Path: "feed.db"},
	Telegraph: telegraphConfig{
		Account:   1,
		ShortName: "Portier",
		Author:    "Portier",
		AuthorURL: "https://github.com/TechMinerApps/portier",
	},
}

// Config is the configuration used in viper
type Config struct {
	DB        dbConfig
	Telegram  bot.Config
	Template  string
	Log       logConfig
	BuntDB    buntDBConfig
	Telegraph telegraphConfig
}

type logConfig struct {
	Mode string
	Path string
}

type dbConfig struct {
	Type     string
	Path     string
	Username string
	Password string
	Host     string
	Port     int
	DBName   string
}
type buntDBConfig struct {
	Path string
}
type telegraphConfig struct {
	Account   int
	ShortName string
	Author    string
	AuthorURL string
}
