package bot

import "gopkg.in/tucnak/telebot.v2"

type App struct {
	Bot *telebot.Bot
}

func (app *App) configCommands() {
	app.Bot.Handle("/start", app.cmdStart)
}
