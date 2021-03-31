package bot

import "gopkg.in/tucnak/telebot.v2"

func (b *bot) cmdHelp(m *telebot.Message) {
	b.app.Logger().Infof("Recieved /help commmand from user: \"%s\"", m.Sender.Username)
	var message = "*Help Message for Portier Feed Bot*\n\n" +
		"/sub \\[URL\\]: subscribe a url\n" +
		"/unsub \\[ID or URL\\]: unsubscribe a feed using id or url\\. ID can be gotten through /list\n" +
		"/list : get current feed list\n" +
		"/help : get this help"

	if _, err := b.bot.Send(m.Chat, message, &telebot.SendOptions{
		DisableWebPagePreview: false,
		DisableNotification:   false,
		ParseMode:             telebot.ModeMarkdownV2,
	}); err != nil {
		b.app.Logger().Errorf("Error sending message: %s", err.Error())
	}
}
