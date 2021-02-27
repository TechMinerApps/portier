package models

// Content is a single passage crawled from feed
type Content struct {
	ID           uint64 `gorm:"primaryKey"`
	HashID       string `gorm:"primaryKey"`
	SourceURL    string
	Title        string
	FeedID       uint64
	Description  string `gorm:"-"` //ignore to db
	TelegraphURL string
}
