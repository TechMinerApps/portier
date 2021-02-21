package bot

import "gopkg.in/tucnak/telebot.v2"

type Bot interface {
	Start()
	Stop()
}
type bot struct {
	Bot *telebot.Bot
}

func NewBot() (Bot, error) {
	var app bot
	var err error
	app.Bot, err = telebot.NewBot(telebot.Settings{})
	if err != nil {
		return nil, err
	}

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
