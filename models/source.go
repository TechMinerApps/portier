package models

type Source struct {
	ID             uint `gorm:"primaryKey;AUTO_INCREMENT"`
	URL            string
	Title          string
	UpdateInterval uint
	ErrorCount     uint
}
