package feed

import (
	"strconv"

	"github.com/TechMinerApps/portier/models"
	"gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

// BroadCaster receive item from channel and broadcast it to any user subscribe to it
type BroadCaster interface {
	Start()
}

type broadcaster struct {
	db          *gorm.DB
	WorkerCount int
	feedChan    <-chan *Feed
	bot         *telebot.Bot
}

type tgRecipient struct {
	ID int64
}

func (t *tgRecipient) Recipient() string {
	return strconv.FormatInt(t.ID, 10)
}

func (b *broadcaster) Start() {
	for i := 0; i < b.WorkerCount; i++ {

		go func() {
			for item := range b.feedChan {
				b.broadcast(item)
			}
		}()
	}
}

func (b *broadcaster) broadcast(item *Feed) {
	var source models.Source
	source.ID = item.SourceID
	var users []*models.User
	b.db.Model(&source).Association("Users")
	b.db.Model(&source).Association("Users").Find(&users)

	for _, u := range users {
		b.send(u.TelegramID, item)
	}

}

func (b *broadcaster) send(telegramID int64, item *Feed) {
	b.bot.Send(&tgRecipient{ID: telegramID}, item.Item.Content)
}
