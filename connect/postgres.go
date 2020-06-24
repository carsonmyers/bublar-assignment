package connect

import (
	"errors"

	// Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"go.uber.org/zap"

	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/jinzhu/gorm"
)

var pgdb *gorm.DB

// Postgres Returns a postgres connection
func Postgres() (*gorm.DB, error) {
	if pgdb != nil {
		if err := pgdb.DB().Ping(); err == nil {
			return pgdb, nil
		}
	}

	config := configure.GetPostgres()
	log.Info("Connecting to postgres", zap.String("config", config.String()))

	db, err := gorm.Open("postgres", config.ConnectionString())
	if err != nil {
		log.Error("Connection to postgres failed", zap.Error(err))
		return nil, errors.New("Postgres connection failed")
	}

	loggerConfig := configure.GetLogger()
	if loggerConfig.Level <= zap.DebugLevel {
		db.LogMode(true)
	}
	db.SingularTable(true)

	pgdb = db
	return pgdb, nil
}
