package feed

import (
	"sync"
	"time"

	"github.com/TechMinerApps/portier/models"
	"github.com/TechMinerApps/portier/modules/log"
	"github.com/TechMinerApps/portier/utils"
	"github.com/mmcdole/gofeed"
	"github.com/tidwall/buntdb"
)

// Poller get feed item from source
// and send it into feedChannel if it is new
type Poller interface {
	Start() error
	Stop() error
	AddSource(s *models.Source) error
	RemoveSource(s *models.Source) error
	FetchTitle(url string) (string, error)
}

type poller struct {
	parser      *gofeed.Parser
	db          *buntdb.DB
	sources     sources
	workers     workers
	feedChannel chan<- *models.Feed
	logger      log.Logger
}

type sources struct {
	Pool []*models.Source
	Lock sync.Mutex
}

type workers struct {
	Pool map[uint]worker
	Lock sync.Mutex
}

type worker struct {
	ticker *time.Ticker
	source models.Source
}

// PollerConfig is the configuration needed to create a Poller
type PollerConfig struct {
	SourcePool  []*models.Source
	DB          *buntdb.DB
	FeedChannel chan<- *models.Feed
	Logger      log.Logger
}

func (p *poller) Start() error {
	// db instance should already have all the info when poller started

	// Protect source Pool
	p.sources.Lock.Lock()
	defer p.sources.Lock.Unlock()
	// Start worker goroutine
	for _, s := range p.sources.Pool {
		go p.worker(s)
		p.logger.Infof("Started poller for %s", s.Title)
	}
	return nil
}

func (p *poller) Stop() error {
	// poller.Stop() does not take care of data persistence

	// Use lock to protect worker pool
	p.workers.Lock.Lock()
	defer p.workers.Lock.Unlock()
	for key, worker := range p.workers.Pool {
		worker.ticker.Stop()
		delete(p.workers.Pool, key)
	}

	// We send feed into this channel
	// so is responsible for closing it
	close(p.feedChannel)
	return nil
}

func (p *poller) AddSource(s *models.Source) error {
	// Protect source Pool
	p.sources.Lock.Lock()
	defer p.sources.Lock.Unlock()
	p.sources.Pool = append(p.sources.Pool, s)
	go p.worker(s)
	return nil
}

func (p *poller) RemoveSource(s *models.Source) error {
	p.workers.Pool[s.ID].ticker.Stop()
	return nil
}

func (p *poller) UpdateSource(s *models.Source) error {
	if p.sources.Pool[s.ID].UpdateInterval <= s.UpdateInterval {
		return nil
	}

	// Stop current worker
	p.workers.Lock.Lock()
	p.workers.Pool[s.ID].ticker.Stop()
	delete(p.workers.Pool, s.ID)
	p.workers.Lock.Unlock()

	// Add new worker
	p.sources.Lock.Lock()
	p.sources.Pool[s.ID] = s
	p.sources.Lock.Unlock()
	go p.worker(s)
	return nil
}

func (p *poller) FetchTitle(url string) (string, error) {
	feed, err := p.parser.ParseURL(url)
	if err != nil {
		return "", err
	}
	return feed.Title, nil
}

func (p *poller) worker(s *models.Source) {
	// worker() is a blocking function
	// that create a create a worker object in p.workers.Pool
	// make sure to call it in a new go routine
	// worker() acquire worker lock, so don't acquire it in upper function
	ticker := time.NewTicker(time.Duration(s.UpdateInterval * uint(time.Second)))
	//ticker := time.NewTicker(time.Second)
	var w worker
	w.ticker = ticker

	// Copy source here
	w.source = *s
	p.workers.Lock.Lock()
	p.workers.Pool[s.ID] = w
	p.workers.Lock.Unlock()

	// Stop blocking when ticker stops
	for range ticker.C {
		p.logger.Infof("Polling source %s", w.source.Title)
		go p.poll(&w.source)
	}

}

func (p *poller) poll(s *models.Source) {
	feed, err := p.parser.ParseURL(s.URL)
	if err != nil {
		p.logger.Warnf("Polling feed %s error: %s", s.Title, err.Error())
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
		feed := models.Feed{
			SourceID: s.ID,
			FeedID:   hash,
			Item:     item,
		}
		p.logger.Infof("Sending feed item from %s to broadcaster", s.Title)
		p.feedChannel <- &feed

		// Store feed item into memdb is done by broadcaster
	}
}

// NewPoller creates a Poller according to the Config
func NewPoller(c *PollerConfig) (Poller, error) {
	var p poller
	p.workers.Pool = make(map[uint]worker)
	p.db = c.DB
	p.feedChannel = c.FeedChannel
	p.logger = c.Logger
	p.sources.Pool = c.SourcePool
	p.parser = gofeed.NewParser()
	return &p, nil
}
