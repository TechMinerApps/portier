package feed

import (
	"errors"
	"strconv"

	"github.com/TechMinerApps/portier/models"
	"github.com/TechMinerApps/portier/modules/log"
	"github.com/TechMinerApps/portier/modules/render"
	"github.com/TechMinerApps/portier/modules/telegraph"
	"github.com/tidwall/buntdb"
	"gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

// BroadCaster receive item from channel and broadcast it to any user subscribe to it
type BroadCaster interface {
	Start()
	Stop()
}

// BroadCastConfig is used to config a broadcaster
type BroadCastConfig struct {

	// DB is used to query users to broadcast
	DB *gorm.DB

	// MemDB is used to store chat id
	MemDB *buntdb.DB

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

	// Telegraph is the config of telegraph module
	Telegraph *telegraph.Config
}

type broadcaster struct {
	renderer render.Renderer
	tgph     telegraph.Telegraph
	BroadCastConfig
}

// Use to implement telebot.Recipient interface
type tgRecipient struct {
	ID int64
}

func (t *tgRecipient) Recipient() string {
	return strconv.FormatInt(t.ID, 10)
}

// NewBroadcaster generates new Broadcaster instance from config
func NewBroadcaster(c *BroadCastConfig) (BroadCaster, error) {
	if c.DB == nil ||
		c.MemDB == nil {
		return nil, errors.New("broadcaster config error")
	}
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
	b.tgph, err = telegraph.NewTelegraph(c.Telegraph)
	if err != nil {
		return nil, err
	}
	return b, nil
}
func (b *broadcaster) Start() {

	b.tgph.Start()

	// Create workers according to WorkerCount
	for i := 0; i < b.WorkerCount; i++ {
		go func() {
			for item := range b.FeedChannel {
				b.Logger.Debugf("Broadcasting feed item %s", item.Item.Title)
				b.broadcast(item)

			}
		}()
	}
}
func (b *broadcaster) Stop() {
	// Do nothing now
}

func (b *broadcaster) broadcast(item *models.Feed) {
	var source models.Source
	source.ID = item.SourceID

	var err error
	item.TelegraphURL, err = b.tgph.Publish(item)
	if err != nil {
		return
	}

	// Find users subscribed
	var users []*models.User
	b.DB.Model(&source).Association("Users")
	b.DB.Model(&source).Association("Users").Find(&users)

	for _, u := range users {

		// Send message sequentially
		b.send(u.TelegramID, item)
	}

}

func (b *broadcaster) send(telegramID int64, item *models.Feed) {

	var err error

	// Render message
	message, err := b.renderer.Render(item)
	if err != nil {
		b.Logger.Errorf("Error rendering message: %s", err.Error())
		return
	}

	// Set telebot options
	options := &telebot.SendOptions{
		DisableWebPagePreview: false,
		ParseMode:             telebot.ModeMarkdownV2,
		DisableNotification:   true,
	}

	// Send via bot
	m, err := b.Bot.Send(&tgRecipient{ID: telegramID}, message, options)
	if err != nil {
		b.Logger.Errorf("Error sending message: %s\n Message is: %s", err.Error(), message)
	}

	// Store the message ID into DB
	// For /unsub to use
	err = b.MemDB.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(strconv.Itoa(m.ID), strconv.Itoa(int(item.SourceID)), nil)
		return err
	})
	if err != nil {
		b.Logger.Errorf("Memory DB insertion error: %s", err.Error())
	}

}
