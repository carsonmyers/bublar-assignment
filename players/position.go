package players

import (
	"fmt"
	"time"

	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

const exp = 48 * time.Hour

// Travel - move a player to a new location
func Travel(player *data.Player, location string) (*data.Position, error) {
	log.Debug("Travel player to new location", zap.String("username", player.Username), zap.String("location", location))
	db, err := connect.Redis()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	pos := &data.Position{}

	key := fmt.Sprintf("%s:position", player.Username)
	encoded, err := db.Get(key).Result()
	if err != nil {
		if err != redis.Nil {
			log.Error("Failed to get position for player", zap.String("username", player.Username), zap.Error(err))
			return nil, errors.EDatabase.NewError(err)
		}
	} else {
		if err := pos.Decode(encoded); err != nil {
			log.Error("Failed to decode player position", zap.String("username", player.Username), zap.String("position", encoded), zap.Error(err))
			return nil, errors.EDatabase.NewError(err)
		}

		key = fmt.Sprintf("location:%s", pos.Location)
		_, err = db.SRem(key, player.Username).Result()
		if err != nil {
			log.Error("Failed to remove player from origin", zap.String("username", player.Username), zap.String("location", pos.Location), zap.Error(err))
			return nil, errors.EDatabase.NewError(err)
		}
	}

	pos.Location = location
	pos.X = 0
	pos.Y = 0

	key = fmt.Sprintf("%s:position", player.Username)
	log.Debug("Storing new position", zap.String("key", key), zap.String("data", pos.Encode()))

	_, err = db.Set(key, pos.Encode(), exp).Result()
	if err != nil {
		log.Error("Failed to set position for player", zap.String("username", player.Username), zap.String("location", location), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	player.Position = pos

	key = fmt.Sprintf("location:%s", location)
	_, err = db.SAdd(key, player.Encode()).Result()
	if err != nil {
		log.Error("Failed to add player to location", zap.String("username", player.Username), zap.String("location", location), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	return pos, nil
}

// Move - set the position of a playwer within their location
func Move(player *data.Player, x int, y int) error {
	db, err := connect.Redis()
	if err != nil {
		return errors.EDatabaseConnection.NewError(err)
	}

	key := fmt.Sprintf("%s:position", player.Username)
	encoded, err := db.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.ENotInLocation.NewErrorf("User is not in a location")
		}

		log.Error("Failed to get position for player", zap.String("username", player.Username), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	pos := &data.Position{}
	if err := pos.Decode(encoded); err != nil {
		log.Error("Failed to decode player position", zap.String("username", player.Username), zap.String("position", encoded), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	player.Position = pos

	posKey := fmt.Sprintf("location:%s", pos.Location)
	if _, err := db.SRem(posKey, player.Encode()).Result(); err != nil {
		log.Error("Failed to remove outdated record from location", zap.String("username", player.Username), zap.String("location", pos.Location), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	pos.X = x
	pos.Y = y

	if _, err := db.Set(key, pos.Encode(), exp).Result(); err != nil {
		log.Error("Failed to set position for player", zap.String("username", player.Username), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	if _, err := db.SAdd(posKey, player.Encode()).Result(); err != nil {
		log.Error("Failed to add location record for moved player", zap.String("username", player.Username), zap.String("location", pos.Location), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	return nil
}
