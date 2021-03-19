package bot

import (
	"errors"
	"net/http"
	"time"

	"github.com/TechMinerApps/portier/modules/feed"
	"github.com/TechMinerApps/portier/modules/log"
	"github.com/tidwall/buntdb"
	"gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

// Config is a config bot used
type Config struct {
	Token string
	MemDB *buntdb.DB
}

// Bot is the control interface provided to portier main instance
type Bot interface {

	// Start is used to start the bot
	Start()

	// Stop stops the bot
	Stop()

	// Bot return the original telebot.Bot object
	Bot() *telebot.Bot
}

// Portier interface is used to communicate to main instance
type Portier interface {
	Poller() feed.Poller
	Logger() log.Logger
	DB() *gorm.DB
}

type bot struct {
	app   Portier
	bot   *telebot.Bot
	memdb *buntdb.DB
}

// NewBot create a bot according to config
func NewBot(c *Config, app Portier) (Bot, error) {
	if c.MemDB == nil {
		return nil, errors.New("memory db is nil, maybe not initialized")
	}
	var err error
	b := &bot{
		app:   app,
		bot:   &telebot.Bot{},
		memdb: c.MemDB,
	}
	b.bot, err = telebot.NewBot(telebot.Settings{
		URL:         "",
		Token:       c.Token,
		Updates:     0,
		Poller:      &telebot.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
		Verbose:     false,
		ParseMode:   "",
		Reporter: func(error) {
		},
		Client: &http.Client{},
	})
	if err != nil {
		return nil, err
	}

	b.configCommands()
	return b, nil
}

func (b *bot) Start() {
	b.bot.Start()
}

func (b *bot) Stop() {
	b.bot.Stop()
}
func (b *bot) Bot() *telebot.Bot {
	return b.bot
}

func (b *bot) configCommands() {
	b.bot.Handle("/start", b.cmdStart)
	b.bot.Handle("/sub", b.cmdSub)
	b.bot.Handle("/unsub", b.cmdUnSub)
	b.bot.Handle("/list", b.cmdList)
}
