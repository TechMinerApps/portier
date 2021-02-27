package feed

import "github.com/mmcdole/gofeed"

// Feed is a struct used in communication between Poller and other modules
type Feed struct {
	FeedID string
	Item   *gofeed.Item
}
