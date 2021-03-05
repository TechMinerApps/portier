package app

import (
	"strings"

	"github.com/TechMinerApps/portier/modules/bot"
	"github.com/TechMinerApps/portier/modules/log"

	"github.com/TechMinerApps/portier/modules/database"
	"github.com/TechMinerApps/portier/modules/feed"
	"github.com/TechMinerApps/portier/utils"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tidwall/buntdb"
	"gorm.io/gorm"
)

type Portier struct {
	db     *gorm.DB
	memDB  *buntdb.DB
	poller feed.Poller
	bot    bot.Bot
	viper  *viper.Viper
	logger log.Logger
	config Config
}

type Config struct {
	DB       database.DBConfig
	Telegram bot.Config
}

func NewPortier() *Portier {
	var p Portier
	p.setupLogger()
	p.setupViper()
	p.setupDB(&p.config.DB)
	p.setupBuntDB()
	p.setupBot()
	return &p

}

func (p *Portier) Start() {
	p.bot.Start()
}

func (p *Portier) setupLogger() error {
	var err error

	p.logger, err = log.NewLogger(&log.Config{
		Mode:       log.HUMAN,
		OutputFile: "",
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Portier) setupDB(c *database.DBConfig) error {
	var err error
	p.db, err = database.NewDBConnection(c)
	if err != nil {
		return err
	}
	return nil
}

func (p *Portier) setupBuntDB() error {
	var err error
	p.memDB, err = buntdb.Open(":memory:")
	if err != nil {
		return err
	}
	return nil
}

func (p *Portier) setupPoller() error {
	return nil
}

func (p *Portier) setupViper() {
	p.viper = viper.New()
	pflag.String("config", "config", "config file name")
	pflag.Parse()
	p.viper.BindPFlags(pflag.CommandLine)

	if p.viper.IsSet("config") {
		p.viper.SetConfigFile(p.viper.GetString("config"))
	} else {
		p.viper.SetConfigName("config")
		p.viper.SetConfigType("yaml")
		p.viper.AddConfigPath(utils.AbsPath(""))
		p.viper.AddConfigPath("/etc/portier")
	}

	p.viper.SetEnvPrefix("PORTIER")
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	p.viper.AutomaticEnv()

	if err := p.viper.ReadInConfig(); err != nil {
		// Used logger here, so setupLogger before setupViper
		p.logger.Fatalf("Unable to read in config: %v", err)
	}

	if err := p.viper.Unmarshal(&p.config); err != nil {
		p.logger.Fatalf("Unable to decode into struct: %v", err)
	}
}

func (p *Portier) setupBot() error {
	var err error
	p.bot, err = bot.NewBot(&p.config.Telegram, p.logger)
	if err != nil {
		return err
	}
	return nil
}
