package models

type User struct {
	ID         int64 `gorm:"primaryKey"`
	TelegramID int64
	Sources    []Source
}
