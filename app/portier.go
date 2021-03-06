package app

import (
	"os"
	"strings"
	"sync"

	"github.com/TechMinerApps/portier/models"
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

// Portier is the main app
type Portier struct {
	db          *gorm.DB
	memDB       *buntdb.DB
	poller      feed.Poller
	broadcaster feed.BroadCaster
	bot         bot.Bot
	viper       *viper.Viper
	logger      log.Logger
	config      Config
	wg          sync.WaitGroup
}

// Config is the configuration used in viper
type Config struct {
	DB       database.DBConfig
	Telegram bot.Config
}

// NewPortier create a new portier object
// does not need config as parameter since this is the main object
func NewPortier() *Portier {
	var p Portier
	p.setupLogger()
	p.setupViper()
	p.setupDB(&p.config.DB)
	p.setupBuntDB()
	p.setupBot()
	p.setupFeedComponent()

	p.logger.Infof("Portier Setup Succeeded")
	return &p

}

// Start is used to start portier instance
// do not return error since error handling show be done within portier object
func (p *Portier) Start() {

	// telebot.Bot.Start() is a blocking method, so start the bot in a goroutine
	go p.bot.Start()
	p.logger.Infof("Telegram Bot Started")

	// Start poller
	p.poller.Start()
	p.logger.Infof("Feed poller started")

	// Start Broadcaster
	p.broadcaster.Start()
	p.logger.Infof("Broadcaster started")

	// Add waitgroup
	p.wg.Add(1)

	p.logger.Infof("Portier started")
}

func (p *Portier) Stop(sig ...os.Signal) {
	if len(sig) != 0 {
	}
	p.wg.Done()
}
func (p *Portier) Wait() {
	p.wg.Wait()
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
	p.db.AutoMigrate(&models.User{}, &models.Source{})
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

func (p *Portier) setupFeedComponent() error {
	feedChan := make(chan *feed.Feed, 10) // hardcoded 10 buffer space
	var sourcePool []*models.Source
	p.db.Model(&models.Source{}).Find(&sourcePool)
	pollerConfig := &feed.PollerConfig{
		SourcePool:  sourcePool,
		DB:          p.memDB,
		FeedChannel: feedChan,
		Logger:      p.logger,
	}
	p.poller, _ = feed.NewPoller(pollerConfig)
	broadcasterConfig := &feed.BroadCastConfig{
		DB:          p.db,
		WorkerCount: 1,
		FeedChannel: feedChan,
		Bot:         p.bot.Bot(),
		Logger:      p.logger,
	}
	p.broadcaster, _ = feed.NewBroadcaster(broadcasterConfig)

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
	p.bot, err = bot.NewBot(&p.config.Telegram, p.logger, p.db)
	if err != nil {
		return err
	}
	return nil
}
