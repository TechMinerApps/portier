package bot

import "gopkg.in/tucnak/telebot.v2"

func (b *bot) cmdStart(m *telebot.Message) {
	b.Bot.Send(m.Chat, "Not Implemented")
}
