package bot

import (
	"strconv"

	"github.com/TechMinerApps/portier/models"
	"github.com/tidwall/buntdb"

	"gorm.io/gorm"

	"gopkg.in/tucnak/telebot.v2"
)

func (b *bot) cmdSub(m *telebot.Message) {
	b.app.Logger().Infof("Recieved /sub commmand from user: \"%s\"", m.Sender.Username)
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
	b.app.Logger().Infof("Add feed \"%s\" to user \"%s\" success", source.Title, m.Sender.Username)
	b.Bot().Send(m.Chat, "Add Feed \""+source.Title+"\" Success")

}
func (b *bot) cmdUnSub(m *telebot.Message) {
	b.app.Logger().Infof("Recieved /unsub commmand from user: \"%s\"", m.Sender.Username)
	var user models.User
	b.app.DB().Model(&user).First(&user).Where("telegram_id = ?", m.Chat.ID)
	if m.IsReply() {
		err := b.memdb.View(func(tx *buntdb.Tx) error {
			val, ok := tx.Get(strconv.Itoa(m.ReplyTo.ID))
			sourceID, err := strconv.Atoi(val)
			if err != nil {
				b.app.Logger().Panicf("Unexpected string to interger convertion error: %s", err.Error())
			}
			var source models.Source
			source.ID = uint(sourceID)
			b.app.DB().Model(source).Association("Users")
			b.app.DB().Model(source).Association("Users").Delete(user)
			return ok
		})
		if err != nil {
			b.app.Logger().Errorf("Memory DB error: %s", err.Error())
			b.Bot().Send(m.Chat, "Database error")
		}
	} else {
		sourceID, err := strconv.Atoi(m.Payload)
		if err != nil {
			b.app.Logger().Infof("/unsub command received illegal input: %s", m.Payload)
			b.Bot().Send(m.Chat, "source ID illegal")
			return
		}
		var source models.Source
		source.ID = uint(sourceID)
		b.app.DB().Model(source).Association("Users")
		if err = b.app.DB().Model(source).Association("Users").Delete(user); err != nil {
			b.app.Logger().Errorf("Database error: %s", err.Error())
			b.Bot().Send(m.Chat, "Database error")
			return
		}

	}
	b.Bot().Send(m.Chat, "Subscription deleted if exist.")

}

func (b *bot) cmdList(m *telebot.Message) {
	b.app.Logger().Infof("Recieved /list commmand from user: \"%s\"", m.Sender.Username)
	var user models.User
	if err := b.app.DB().Model(user).First(&user).Where("telegram_id = ?", m.Chat.ID).Error; err != nil {
		b.app.Logger().Errorf("Database error: %s", err.Error())
		b.Bot().Send(m.Chat, "Database error")
		return
	}
	var sources []models.Source
	b.app.DB().Model(user).Association("Sources")
	b.app.DB().Model(user).Association("Sources").Find(&sources)
	var message string
	if len(sources) == 0 {
		b.bot.Send(m.Chat, "No subscription")
		return
	}
	for _, s := range sources {
		message = message + "[" + strconv.Itoa(int(s.ID)) + "]: " + s.Title + "\n"
	}
	b.bot.Send(m.Chat, message)

}
