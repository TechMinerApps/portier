package models

import "github.com/mmcdole/gofeed"

// Feed is a struct used in communication between Poller and other modules
type Feed struct {
	SourceID uint
	FeedID   string
	Item     *gofeed.Item
}
