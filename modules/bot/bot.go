package bot

import "gopkg.in/tucnak/telebot.v2"

type App struct {
	Bot *telebot.Bot
}

func NewTeleBot() (*App, error) {
	var app App
	var err error
	app.Bot, err = telebot.NewBot(telebot.Settings{})
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (app *App) Start() {
	app.Bot.Start()
}

func (app *App) Stop() {
	app.Bot.Stop()
}

func (app *App) configCommands() {
	app.Bot.Handle("/start", app.cmdStart)
}
