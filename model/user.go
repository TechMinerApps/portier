package model

type User struct {
	ID         int64 `gorm:"primaryKey"`
	TelegramID int64
	Feeds      []Feed `gorm:"many2many:feeds;"`
	State      int    `gorm:"DEFAULT:0;"`
}
