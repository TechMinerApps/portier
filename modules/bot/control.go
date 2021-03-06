package bot

import (
	"github.com/TechMinerApps/portier/models"
	"gopkg.in/tucnak/telebot.v2"
)

func (b *bot) cmdStart(m *telebot.Message) {
	b.logger.Infof("User \"%s\" /start recieved", m.Sender.Username)
	var user models.User
	user.TelegramID = m.Chat.ID
	if err := b.db.Where(&user).FirstOrCreate(&user).Error; err != nil {
		b.logger.Errorf("Database error: %s", err.Error())
	}
	b.logger.Infof("New user \"%s\" registered into database with ID: %d", m.Sender.Username, m.Chat.ID)
	b.Bot.Send(m.Chat, "Welcome to portier")
}
