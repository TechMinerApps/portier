package feed

import (
	"strconv"

	"github.com/TechMinerApps/portier/models"
	"github.com/TechMinerApps/portier/modules/log"
	"github.com/TechMinerApps/portier/modules/render"
	"gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

// BroadCaster receive item from channel and broadcast it to any user subscribe to it
type BroadCaster interface {
	Start()
}

// BroadCastConfig is used to config a broadcaster
type BroadCastConfig struct {

	// DB is used to query users to broadcast
	DB *gorm.DB

	// WorkerCount is used in concurrent broadcast
	WorkerCount int

	// FeedChannel is where feed item comes from
	FeedChannel <-chan *models.Feed

	// Bot is the bot which broadcaster broadcast to
	Bot *telebot.Bot

	// Logger is used to log events
	Logger log.Logger

	// Template is a string used to render text
	Template string
}

type broadcaster struct {
	renderer render.Renderer
	BroadCastConfig
}

type tgRecipient struct {
	ID int64
}

func (t *tgRecipient) Recipient() string {
	return strconv.FormatInt(t.ID, 10)
}

func NewBroadcaster(c *BroadCastConfig) (*broadcaster, error) {
	b := &broadcaster{
		BroadCastConfig: *c,
	}

	var err error
	cfg := render.Config{
		Template: c.Template,
	}
	b.renderer, err = render.NewRenderer(cfg)
	if err != nil {
		return nil, err
	}

	return b, nil
}
func (b *broadcaster) Start() {
	for i := 0; i < b.WorkerCount; i++ {

		go func() {
			for item := range b.FeedChannel {
				b.Logger.Debugf("Broadcasting feed item %s", item.Item.Title)
				b.broadcast(item)
			}
		}()
	}
}

func (b *broadcaster) broadcast(item *models.Feed) {
	var source models.Source
	source.ID = item.SourceID
	var users []*models.User
	b.DB.Model(&source).Association("Users")
	b.DB.Model(&source).Association("Users").Find(&users)

	for _, u := range users {
		b.send(u.TelegramID, item)
	}

}

func (b *broadcaster) send(telegramID int64, item *models.Feed) {
	var err error
	message, err := b.renderer.Render(item)
	if err != nil {
		b.Logger.Errorf("Error sending message: %s", err.Error())
		return
	}
	options := &telebot.SendOptions{
		DisableWebPagePreview: false,
		ParseMode:             telebot.ModeMarkdownV2,
		DisableNotification:   true,
	}
	_, err = b.Bot.Send(&tgRecipient{ID: telegramID}, message, options)
	if err != nil {
		b.Logger.Errorf("Error sending message: %s", err.Error())
	}
}
