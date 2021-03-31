package app

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/TechMinerApps/portier/models"
	"github.com/TechMinerApps/portier/modules/bot"
	"github.com/TechMinerApps/portier/modules/log"
	"github.com/TechMinerApps/portier/modules/telegraph"

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

// NewPortier create a new portier object
// does not need config as parameter since this is the main object
func NewPortier() *Portier {

	var p Portier

	// All the setup* func should handle error by itself

	// Read in config first
	p.setupViper()

	// Logger must be set up before any other setup
	p.setupLogger()

	p.setupDB()
	p.setupBuntDB()
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

// Stop shutdown portier gracefully
// can accept a list of signals, print them if provided
func (p *Portier) Stop(sig ...os.Signal) {

	// Close buntdb
	p.memDB.Close()

	// Close DB
	db, err := p.db.DB()
	if err != nil {
		// Really should not be an error
		p.logger.Panicf("Getting GORM DB instance error")
	}
	db.Close()

	// Debug info
	if len(sig) != 0 {
		p.logger.Debugf("Recieved signal: %v", sig)
	}
	p.logger.Infof("Shutting down")
	p.wg.Done()
}

// Wait is a blocking function that wait for portier to stop
func (p *Portier) Wait() {
	p.wg.Wait()
}

// Poller return the poller portier used
func (p *Portier) Poller() feed.Poller {
	return p.poller
}

// Logger returns the logger portier used
func (p *Portier) Logger() log.Logger {
	return p.logger
}

// DB returns the database instance portier used
func (p *Portier) DB() *gorm.DB {
	return p.db
}

func (p *Portier) setupLogger() {
	var err error

	p.logger, err = log.NewLogger(&log.Config{
		Mode:       log.ConvertToLoggerType(p.config.Log.Mode),
		OutputFile: p.config.Log.Path,
	})
	if err != nil {

		// Fatal error
		fmt.Printf("Error setting up logger: %s\n", err.Error())
		os.Exit(-1)
	}
}

func (p *Portier) setupDB() {
	var err error

	cfg := database.DBConfig{
		Type:     database.ConvertToDBType(p.config.DB.Type),
		Path:     p.config.DB.Path,
		Username: p.config.DB.Username,
		Password: p.config.DB.Password,
		Host:     p.config.DB.Host,
		Port:     p.config.DB.Port,
		DBName:   p.config.DB.DBName,
	}

	// Connect to database
	p.db, err = database.NewDBConnection(&cfg)
	if err != nil {
		p.logger.Fatalf("Error setting up database: %s", err.Error())
	}

	// Create table if not exist
	p.db.AutoMigrate(&models.User{}, &models.Source{})
}

func (p *Portier) setupBuntDB() {
	var err error

	// Create a kv db to store feeds
	// By default, buntdb will do fsync every second
	p.memDB, err = buntdb.Open(p.config.BuntDB.Path)
	if err != nil {
		p.logger.Fatalf("BuntDB error: %s", err.Error())
	}
}

func (p *Portier) setupFeedComponent() {
	var err error
	feedChan := make(chan *models.Feed, 10) // hardcoded 10 buffer space
	var sourcePool []*models.Source

	// Load sources into var sourcePool
	p.db.Model(&models.Source{}).Find(&sourcePool)

	// Setup poller
	pollerConfig := &feed.PollerConfig{
		SourcePool:  sourcePool,
		DB:          p.memDB,
		FeedChannel: feedChan,
		Logger:      p.logger,
	}
	p.poller, err = feed.NewPoller(pollerConfig)
	if err != nil {
		p.logger.Fatalf("Error creating poller: %s", err.Error())
	}

	// Setup Bot here
	// Because bot rely on poller, need poller object to interact with sources
	p.setupBot()

	// Then setup broadcaster
	// Broadcaster rely on bot to broadcast
	broadcasterConfig := &feed.BroadCastConfig{
		DB:          p.db,
		MemDB:       p.memDB,
		WorkerCount: 1,
		FeedChannel: feedChan,
		Bot:         p.bot.Bot(),
		Logger:      p.logger,
		Template:    p.config.Template,
		Telegraph:   &telegraph.Config{AccountNumber: p.config.Telegraph.Account, ShortName: p.config.Telegraph.ShortName, AuthorName: p.config.Telegraph.Author, AuthorURL: p.config.Telegraph.AuthorURL, AccessToken: []string{}, Logger: p.logger},
	}
	p.broadcaster, err = feed.NewBroadcaster(broadcasterConfig)
	if err != nil {
		p.logger.Fatalf("Error creating broadcaster: %s", err.Error())
	}

}

func (p *Portier) setupViper() {
	p.viper = viper.New()

	// Allow --config flag to set config file
	pflag.String("config", "config", "config file name")
	pflag.Parse()
	p.viper.BindPFlags(pflag.CommandLine)

	if p.viper.IsSet("config") {
		p.viper.SetConfigFile(p.viper.GetString("config"))
	} else {

		p.viper.SetConfigName("config")

		// Use YAML as config format
		p.viper.SetConfigType("yaml")

		// Allow ./config.yaml
		p.viper.AddConfigPath(utils.AbsPath(""))

		// Allow /etc/portier/config.yaml
		p.viper.AddConfigPath("/etc/portier")
	}

	// Setup environment variable parsing
	p.viper.SetEnvPrefix("PORTIER")
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	p.viper.AutomaticEnv()

	if err := p.viper.ReadInConfig(); err != nil {

		// Fatal error
		// Do not use logger here, logger is not yet setup
		fmt.Printf("Unable to read in config: %v\n", err)
		os.Exit(-1)
	}

	// Load default config
	p.config = DefaultConfig

	// Load config into p.config
	if err := p.viper.Unmarshal(&p.config); err != nil {
		// Fatal error
		fmt.Printf("Unable to unmarshal into struct: %vi\n", err)
		os.Exit(-1)
	}
}

func (p *Portier) setupBot() {
	var err error
	cfg := bot.Config{
		Token: p.config.Telegram.Token,
		MemDB: p.memDB,
	}
	p.bot, err = bot.NewBot(&cfg, p)
	if err != nil {
		p.logger.Fatalf("Bot setup failed: %s", err.Error())
	}
}
