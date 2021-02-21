package feed

import (
	"time"

	"github.com/TechMinerApps/portier/models"
	"github.com/TechMinerApps/portier/utils"
	"github.com/mmcdole/gofeed"
	"github.com/tidwall/buntdb"
)

// Poller get feed item from source
// and send it into feedChannel if it is new
type Poller interface {
	Start() error
	Stop() error
}

type poller struct {
	parser      gofeed.Parser
	db          *buntdb.DB
	sourcePool  []*models.Source
	feedChannel chan<- *Feed
}

// Config is the configuration needed to create a Poller
type Config struct {
	SourcePool  []*models.Source
	DB          *buntdb.DB
	FeedChannel chan<- *Feed
}

func (p *poller) Start() error {
	// db instance should already have all the info when poller started

	// Start worker goroutine
	for _, s := range p.sourcePool {
		go p.worker(s)
	}
	return nil
}

func (p *poller) Stop() error {
	// poller.Stop() does not take care of data persistence
	return nil
}

func (p *poller) worker(s *models.Source) {
	// worker is a blocking function
	// make sure to call it in a new go routine
	tickerChan := time.Tick(time.Duration(s.UpdateInterval))
	for {
		go p.poll(s)
		<-tickerChan
	}

}

func (p *poller) poll(s *models.Source) {
	feed, err := p.parser.ParseURL(s.URL)
	if err != nil {
		return
	}
	for _, item := range feed.Items {
		err = p.db.View(func(tx *buntdb.Tx) error {

			hash := utils.StringHash(s.URL + "|" + item.GUID)

			// Check if item exists in memory
			_, ok := tx.Get(hash)
			if ok == nil {
				return nil
			} else if ok != buntdb.ErrNotFound { // Report error
				return ok
			}

			// Send item if new
			p.feedChannel <- &Feed{
				FeedID: hash,
				Item:   item,
			}

			// End Transaction
			return nil
		})
	}
}

// NewPoller creates a Poller according to the Config
func NewPoller(c *Config) (Poller, error) {
	var p poller
	return &p, nil
}
