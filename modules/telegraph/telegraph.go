package telegraph

import (
	"sync"
	"time"

	"github.com/TechMinerApps/portier/models"
	tgraph "github.com/meinside/telegraph-go"
)

const (
	ErrTelegraphFlood = iota
)

type Telegraph interface {

	// Publish is a blocking function that insert the provided feed into queue and wait for process
	Publish(item *models.Feed) (string, error)
}

type Config struct {
	AccessToken []string
}

type Item struct {
	ResultChan chan<- string
	Feed       *models.Feed
}

type telegraph struct {
	clientPool    []*tgraph.Client
	currentClient int
	lock          sync.Mutex
	queue         chan *Item
}

func NewTelegraph(c *Config) (Telegraph, error) {
	var t telegraph
	for _, token := range c.AccessToken {
		client, err := tgraph.Load(token)
		if err != nil {
			return nil, err
		}
		t.clientPool = append(t.clientPool, client)
	}
	return &t, nil
}

func (t *telegraph) Start() error {
	go func() {
		for item := range t.queue {
			for {
				url, err := t.publish(item)
				if err == nil {
					item.ResultChan <- url
					break
				} else {
					time.Sleep(60 * time.Second)
				}
			}

		}
	}()
	return nil
}

func (t *telegraph) Publish(feed *models.Feed) (string, error) {
	resultCh := make(chan string)
	item := &Item{
		ResultChan: resultCh,
		Feed:       feed,
	}

	// Send item into queue
	t.queue <- item

	// Wait for result

	url := <-resultCh

	return url, nil
}

func (t *telegraph) publish(item *Item) (string, error) {

	// publish should not mess with channel

	htmlContent := item.Feed.Item.Content
	var err error = nil
	var url string = ""
	t.lock.Lock()
	if page, err1 := t.clientPool[t.currentClient].CreatePageWithHTML(
		item.Feed.Item.Title,
		item.Feed.Item.Author.Name,
		item.Feed.Item.Link,
		htmlContent,
		true); err1 == nil {
		url = page.URL

	} else {
		err = err1
	}
	t.lock.Unlock()
	return url, err
}
