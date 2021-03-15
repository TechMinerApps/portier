package telegraph

import (
	"errors"
	"sync"
	"time"

	"github.com/TechMinerApps/portier/models"
	"github.com/TechMinerApps/portier/modules/log"
	tgraph "github.com/TechMinerApps/telegraph"
)

type Telegraph interface {

	// Publish is a blocking function that insert the provided feed into queue and wait for process
	Publish(item *models.Feed) (string, error)

	// Start is used to start a instance
	// telegraph do not needs to stop, just close the input channel
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
	clientPool    []tgraph.Client
	currentClient int
	lock          sync.Mutex
	queue         chan *Item
}

func NewTelegraph(c *Config) (Telegraph, error) {
	t := &telegraph{
		logger: c.Logger,
		config: c,

		// Load balance with round-robin
		clientPool:    []tgraph.Client{},
		currentClient: 0,

		// Mutex lock to protect currentClient
		lock:  sync.Mutex{},
		queue: make(chan *Item),
	}

	// If newly created instance
	if len(c.AccessToken) == 0 {
		if err := t.createAccount(); err != nil {
			return nil, err
		}
		return t, nil
	}

	// Spawn clients from access token
	for _, token := range c.AccessToken {
		client, err := tgraph.NewClientWithToken(token)
		if err != nil {
			return nil, err
		}
		t.clientPool = append(t.clientPool, client)
	}
	return t, nil
}

func (t *telegraph) createAccount() error {
	for i := 0; i < t.config.AccountNumber; i++ {
		AccountInfo := tgraph.Account{
			AccessToken: "",
			AuthURL:     "",
			ShortName:   t.config.ShortName,
			AuthorName:  t.config.AuthorName,
			AuthorURL:   t.config.AuthorURL,
			PageCount:   0,
		}
		client := tgraph.NewClient()
		_, err := client.CreateAccount(AccountInfo)
		if err != nil {
			return err
		}
		t.clientPool = append(t.clientPool, client)
		t.config.AccessToken = append(t.config.AccessToken, client.Account().AccessToken)
		t.logger.Infof("Created Telegraph account %d success", i)

		// Avoid flood wait
		// Can lead to long waiting at first time start
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
				} else if err == tgraph.ErrFloodWait {
					t.logger.Warnf("Recieve Telegraph flood wait: %s", err.Error())

					// Flood wait 7s, but wait a longer 10 seconds to ensure success
					time.Sleep(10 * time.Second)
				} else {
					t.logger.Errorf("Error publishing to telegraph: %s", err.Error())
					item.ResultChan <- ""
					break
				}
			}

		}
	}()
}

// Publish is a blocking function that wait for the page to return or return a error
func (t *telegraph) Publish(feed *models.Feed) (string, error) {
	resultCh := make(chan string)
	defer close(resultCh)
	item := &Item{
		ResultChan: resultCh,
		Feed:       feed,
	}

	// Send item into queue
	t.queue <- item

	// Wait for result
	url := <-resultCh
	if url == "" {
		return "", errors.New("receiving empty url, may be error")
	}

	return url, nil
}

func (t *telegraph) publish(item *Item) (string, error) {

	// publish should not mess with channel

	// content directly from feed
	htmlContent := item.Feed.Item.Content
	if htmlContent == "" {
		// Must have content
		htmlContent = "Empty Content"
	}

	// use mutex lock to avoid concurrent use of same client
	t.lock.Lock()
	defer t.lock.Unlock()
	defer func() {
		t.currentClient++
		if t.currentClient == len(t.clientPool) {
			t.currentClient = 0
		}
	}()

	// Create a new page
	content, err := t.clientPool[t.currentClient].ContentFormat(htmlContent)
	if err != nil {
		return "", err
	}
	page := tgraph.Page{
		Title:       item.Feed.Item.Title,
		Description: item.Feed.Item.Description,
		AuthorName:  item.Feed.Item.Author.Name,
		AuthorURL:   item.Feed.Item.Link,
		Content:     content,
	}
	if page, err := t.clientPool[t.currentClient].CreatePage(page, false); err != nil {
		return "", err
	} else {
		return page.URL, nil
	}
}
