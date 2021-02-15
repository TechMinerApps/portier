package bot

import "gopkg.in/tucnak/telebot.v2"

func (app *App) cmdStart(m *telebot.Message) {
	app.Bot.Send(m.Chat, "Not Implemented")
}
