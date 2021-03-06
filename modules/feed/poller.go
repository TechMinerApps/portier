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
	workerPool  map[uint]worker
	feedChannel chan<- *Feed
}

type worker struct {
	ticker *time.Ticker
	source models.Source
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
	for _, worker := range p.workerPool {
		worker.ticker.Stop()
	}
	close(p.feedChannel)
	return nil
}

func (p *poller) worker(s *models.Source) {
	// worker is a blocking function
	// that create a create a worker object in p.workerPool
	// make sure to call it in a new go routine
	ticker := time.NewTicker(time.Duration(s.UpdateInterval))
	var w worker
	w.ticker = ticker
	p.workerPool[s.ID] = w
	for {
		go p.poll(&w.source)
		<-ticker.C
	}

}

func (p *poller) poll(s *models.Source) {
	feed, err := p.parser.ParseURL(s.URL)
	if err != nil {
		// Error Handling needed
		return
	}
	for _, item := range feed.Items {
		hash := utils.StringHash(s.URL + "|" + item.GUID)

		// Do a read transaction to check if feed exists
		err = p.db.View(func(tx *buntdb.Tx) error {
			// Check if item exists in memory
			_, ok := tx.Get(hash)
			return ok
		})

		if err == nil {
			// End if found without error
			break
		} else if err != buntdb.ErrNotFound {
			// Report Error
			break
		}

		// Send item if new
		feed := Feed{
			SourceID: s.ID,
			FeedID:   hash,
			Item:     item,
		}
		p.feedChannel <- &feed

		// Then store it in db
		err = p.db.Update(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set(hash, "exists", nil)
			return err
		})
	}
}

// NewPoller creates a Poller according to the Config
func NewPoller(c *Config) (Poller, error) {
	var p poller
	p.workerPool = make(map[uint]worker)
	return &p, nil
}
