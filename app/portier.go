package app

import (
	"github.com/TechMinerApps/portier/modules/bot"
	"github.com/TechMinerApps/portier/modules/database"
	"github.com/TechMinerApps/portier/modules/feed"
	"github.com/tidwall/buntdb"
	"gorm.io/gorm"
)

type Portier struct {
	db     *gorm.DB
	memDB  *buntdb.DB
	poller feed.Poller
	bot    bot.Bot
}

type Config struct {
	DB database.DBConfig
}

func NewPortier(c *Config) (*Portier, error) {
	var p Portier
	p.setupDB(&c.DB)
	p.setupBuntDB()
	return &p, nil

}

func (p *Portier) setupDB(c *database.DBConfig) error {
	var err error
	p.db, err = database.NewDBConnection(c)
	if err != nil {
		return err
	}
	return nil
}

func (p *Portier) setupBuntDB() error {
	var err error
	p.memDB, err = buntdb.Open(":memory:")
	if err != nil {
		return err
	}
	return nil
}

func (p *Portier) setupPoller() error {
	return nil
}
