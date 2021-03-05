package bot

import (
	"net/http"
	"time"

	"github.com/TechMinerApps/portier/modules/log"
	"gopkg.in/tucnak/telebot.v2"
)

type Config struct {
	Token string
}

type Bot interface {
	Start()
	Stop()
}
type bot struct {
	Bot    *telebot.Bot
	logger log.Logger
}

func NewBot(c *Config, logger log.Logger) (Bot, error) {
	var app bot
	var err error
	app.Bot, err = telebot.NewBot(telebot.Settings{
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
	app.logger = logger
	return &app, nil
}

func (b *bot) Start() {
	b.Bot.Start()
}

func (b *bot) Stop() {
	b.Bot.Stop()
}

func (b *bot) configCommands() {
	b.Bot.Handle("/start", b.cmdStart)
}
