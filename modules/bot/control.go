package bot

import (
	"github.com/TechMinerApps/portier/models"
	"gopkg.in/tucnak/telebot.v2"
)

func (b *bot) cmdStart(m *telebot.Message) {
	b.app.Logger().Infof("User \"%s\" /start recieved", m.Sender.Username)
	var user models.User
	user.TelegramID = m.Chat.ID
	if err := b.app.DB().Where(&user).FirstOrCreate(&user).Error; err != nil {
		b.app.Logger().Errorf("Database error: %s", err.Error())
	}
	b.app.Logger().Infof("New user \"%s\" registered into database with ID: %d", m.Sender.Username, m.Chat.ID)
	b.Bot().Send(m.Chat, "Welcome to portier\nuse /help to check usage")
}
