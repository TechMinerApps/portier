package telegraph

import (
	"sync"
	"time"

	"github.com/TechMinerApps/portier/models"
	"github.com/TechMinerApps/portier/modules/log"
	tgraph "github.com/meinside/telegraph-go"
)

const (
	ErrTelegraphFlood = iota
)

type Telegraph interface {

	// Publish is a blocking function that insert the provided feed into queue and wait for process
	Publish(item *models.Feed) (string, error)

	Start()
}

type Config struct {
	AccountNumber int
	ShortName     string
	AuthorName    string
	AuthorURL     string

	AccessToken []string
	Logger      log.Logger
}

type Item struct {
	ResultChan chan<- string
	Feed       *models.Feed
}

type telegraph struct {
	logger        log.Logger
	config        *Config
	clientPool    []*tgraph.Client
	currentClient int
	lock          sync.Mutex
	queue         chan *Item
}

func NewTelegraph(c *Config) (Telegraph, error) {
	t := &telegraph{
		logger:        c.Logger,
		config:        c,
		clientPool:    []*tgraph.Client{},
		currentClient: 0,
		lock:          sync.Mutex{},
		queue:         make(chan *Item),
	}

	if len(c.AccessToken) == 0 {
		if err := t.createAccount(); err != nil {
			return nil, err
		}
		return t, nil
	}
	for _, token := range c.AccessToken {
		client, err := tgraph.Load(token)
		if err != nil {
			return nil, err
		}
		t.clientPool = append(t.clientPool, client)
	}
	return t, nil
}

func (t *telegraph) createAccount() error {
	for i := 0; i < t.config.AccountNumber; i++ {
		client, err := tgraph.Create(t.config.ShortName, t.config.AuthorName, t.config.AuthorURL)
		if err != nil {
			return err
		}
		t.clientPool = append(t.clientPool, client)
		t.config.AccessToken = append(t.config.AccessToken, client.AccessToken)
		t.logger.Infof("Created Telegraph account %d success", i)
		time.Sleep(time.Second)
	}
	return nil
}

func (t *telegraph) Start() {
	go func() {
		for item := range t.queue {
			for {
				url, err := t.publish(item)
				if err == nil {
					item.ResultChan <- url
					break
				} else {
					t.logger.Warnf("Error publishing to telegraph: %s", err.Error())
					time.Sleep(60 * time.Second)
				}
			}

		}
	}()
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
	close(resultCh)

	return url, nil
}

func (t *telegraph) publish(item *Item) (string, error) {

	// publish should not mess with channel

	htmlContent := item.Feed.Item.Content
	if htmlContent == "" {
		htmlContent = "Empty Content"
	}
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
	t.currentClient++
	if t.currentClient == len(t.clientPool) {
		t.currentClient = 0
	}
	t.lock.Unlock()
	return url, err
}
