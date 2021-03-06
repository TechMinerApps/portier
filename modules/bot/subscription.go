package bot

import (
	"github.com/TechMinerApps/portier/models"
	"gopkg.in/tucnak/telebot.v2"
)

func (b *bot) cmdSub(m *telebot.Message) {
	var source models.Source
	source.URL, _ = GetURLAndMentionFromMessage(m)
	source.UpdateInterval = 300 // hardcoded for now
	var user models.User
	b.db.Model(&user).Where("telegram_id = ?", m.Chat.ID).Association("Sources").Append(&source)
	b.Bot.Send(m.Chat, "Success")

}
