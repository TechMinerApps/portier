package database

import (
	"strconv"

	"github.com/TechMinerApps/portier/utils"
	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBType is just a renamed int
type DBType int

// DBType constants
const (
	SQLITE DBType = iota
	MYSQL
)

// DBConfig is the config used to start a DB connection
type DBConfig struct {
	Type         DBType
	SQLiteConfig sqliteConfig `mapstructure:",squash"`
	MySQLConfig  mysqlConfig  `mapstructure:",squash"`
}
type sqliteConfig struct {
	Path string
}

type mysqlConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	DBName   string
}

// NewDBConnection returns a DB object based on config provided
func NewDBConnection(c *DBConfig) (*gorm.DB, error) {

	var err error
	var DB *gorm.DB

	switch c.Type {
	case SQLITE:
		path := utils.AbsPath(c.SQLiteConfig.Path)
		DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	case MYSQL:
		cfg := mysqldriver.NewConfig()
		cfg.User = c.MySQLConfig.Username
		cfg.Passwd = c.MySQLConfig.Password
		cfg.Net = "tcp"
		cfg.Addr = c.MySQLConfig.Host + ":" + strconv.Itoa(c.MySQLConfig.Port)
		cfg.DBName = c.MySQLConfig.DBName
		// Charset is utf8mb4 by default
		DB, err = gorm.Open(mysql.New(mysql.Config{
			DSN: cfg.FormatDSN(),
		}), &gorm.Config{})
	}

	// Handle errors
	if err != nil {
		return nil, err
	}
	if DB.Error != nil {
		return nil, DB.Error
	}

	return DB, nil
}
