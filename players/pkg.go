package players

import (
	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/carsonmyers/bublar-assignment/logger"
)

var (
	log    = logger.GetLogger()
	models = make([]interface{}, 0, 5)
)

// Migrate - migrate all tables managed by this service
func Migrate() error {
	db, err := connect.Postgres()
	if err != nil {
		return errors.EDatabaseConnection.NewError(err)
	}

	db.AutoMigrate(models...)
	return nil
}
