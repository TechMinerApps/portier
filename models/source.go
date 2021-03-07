package models

type Source struct {
	ID             uint    `gorm:"primaryKey;AUTO_INCREMENT"`
	Users          []*User `gorm:"many2many:user_sources"`
	URL            string
	Title          string
	UpdateInterval uint
	ErrorCount     uint
}
