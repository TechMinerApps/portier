package bot

import (
	"net/http"
	"time"

	"github.com/TechMinerApps/portier/modules/feed"
	"github.com/TechMinerApps/portier/modules/log"
	"gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

type Config struct {
	Token string
}

type Bot interface {

	// Start is used to start the bot
	Start()

	// Stop stops the bot
	Stop()

	// Bot return the original telebot.Bot object
	Bot() *telebot.Bot
}
type bot struct {
	poller feed.Poller
	bot    *telebot.Bot
	db     *gorm.DB
	logger log.Logger
}

func NewBot(c *Config, logger log.Logger, db *gorm.DB, poller feed.Poller) (Bot, error) {
	var app bot
	var err error
	app.bot, err = telebot.NewBot(telebot.Settings{
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

	app.configCommands()
	app.poller = poller
	app.logger = logger
	app.db = db
	return &app, nil
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
}
