package bot

import (
	"github.com/TechMinerApps/portier/models"

	"gorm.io/gorm"

	"gopkg.in/tucnak/telebot.v2"
)

func (b *bot) cmdSub(m *telebot.Message) {
	var source models.Source
	source.URL, _ = GetURLAndMentionFromMessage(m)
	source.Title, _ = b.app.Poller().FetchTitle(source.URL)
	source.UpdateInterval = 300 // hardcoded for now
	var user models.User
	if err := b.app.DB().Model(&user).Association("Sources").Error; err != nil {
		b.app.Logger().Errorf("Error starting association mode: %s", err.Error())
		b.Bot().Send(m.Chat, "Database error")
		return
	}
	if err := b.app.DB().Model(&user).Where("telegram_id = ?", m.Chat.ID).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			b.app.Logger().Errorf("Database error: %s", err.Error())
			b.Bot().Send(m.Chat, "Database error")
			return
		}
		b.app.Logger().Errorf("Chat ID not registered")
		b.Bot().Send(m.Chat, "Chat ID not registered, please run /start first")
		return
	}
	b.app.DB().Model(&source).Where("url = ?", source.URL).First(&source)
	b.app.DB().Model(&user).Association("Sources").Append(&source)

	b.app.Poller().AddSource(&source)
	b.Bot().Send(m.Chat, "Add Feed \""+source.Title+"\" Success")

}
