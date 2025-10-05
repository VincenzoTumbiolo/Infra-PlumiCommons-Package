package utility

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const _DSN_TEMPLATE = "host=%s user=%s password=%s dbname=%s port=%d"

var _dsn string
var _DB *gorm.DB

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

func DB() *gorm.DB {
	if _DB != nil {
		return _DB
	}

	db, err := gorm.Open(postgres.Open(_dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Couldn't connect to db", "err", err)
		panic(err)
	}

	_DB = db

	return db
}

func InitModule(dbConnectionDetails DBConfig) {
	_dsn = fmt.Sprintf(
		_DSN_TEMPLATE,
		dbConnectionDetails.Host,
		dbConnectionDetails.User,
		dbConnectionDetails.Password,
		dbConnectionDetails.Database,
		dbConnectionDetails.Port,
	)
}
