package models

// Feed is a RSS/Atom Feed
type Feed struct {
	ID         uint `gorm:"primary_key;AUTO_INCREMENT"`
	URL        string
	Title      string
	ErrorCount uint
	Content    []Content
}
