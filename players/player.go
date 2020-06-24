package players

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gbrlsnchs/jwt/v2"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"

	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/carsonmyers/bublar-assignment/errors"
	"go.uber.org/zap"
)

// Player - persisted user data
type Player struct {
	Username  string    `json:"username" gorm:"primary_key"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:timestamp"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"type:timestamp"`
}

// ToPlayer - convert to universal data format
func (p *Player) ToPlayer() *data.Player {
	pw := p.Password

	return &data.Player{
		Username: p.Username,
		Password: &pw,
	}
}

// CreatePlayer - add a new player to the system
func CreatePlayer(player *data.Player) (*data.Player, error) {
	db, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var count uint64
	q := db.Model(&Player{}).Where(&Player{Username: player.Username}).Count(&count)
	if err := q.Error; err != nil {
		log.Error("Error counting existing users", zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	if count != 0 {
		log.Error("Attempt to create a duplicate user", zap.String("username", player.Username))
		return nil, errors.EDuplicateUser.NewError(player.Username)
	}

	var pw string
	if player.Password != nil {
		pw = *player.Password
	}

	hashed, err := hashPassword(pw)
	if err != nil {
		return nil, errors.EInternal.NewError(err)
	}

	playerModel := &Player{
		Username:  player.Username,
		Password:  hashed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(playerModel).Error; err != nil {
		log.Error("Failed to store new user", zap.String("username", player.Username), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	return playerModel.ToPlayer(), nil
}

// AuthPlayer - athenticate an existing player, generating an auth token
func AuthPlayer(username, password string) (*jwt.JWT, error) {
	db, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var player Player
	if err := db.Where(&Player{Username: username}).First(&player).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.ENotFound.NewError(err)
		}

		log.Error("Error fetching user for auth", zap.String("username", username), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	if !testPassword(password, player.Password) {
		log.Error("Failed player login", zap.String("username", username))
		return nil, errors.EAuth.NewErrorf("login failed")
	}

	return &jwt.JWT{
		Issuer:         "bublar-assignment",
		Subject:        "bublar-player",
		Audience:       username,
		ExpirationTime: time.Now().Add(time.Hour).Unix(),
		IssuedAt:       time.Now().Unix(),
	}, nil
}

// GetPlayer - fetch a single player by username
func GetPlayer(id string) (*data.Player, error) {
	pdb, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var player Player
	if err := pdb.Where(&Player{Username: id}).First(&player).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.ENotFound.NewError(err)
		}

		log.Error("Error fetching user data", zap.String("username", id), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	result := player.ToPlayer()
	rdb, err := connect.Redis()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	key := fmt.Sprintf("%s:position", player.Username)
	encoded, err := rdb.Get(key).Result()
	if err != nil {
		if err != redis.Nil {
			log.Error("Failed to get position for player", zap.String("username", player.Username), zap.Error(err))
			return nil, errors.EDatabase.NewError(err)
		} else {
			log.Debug("No position for user", zap.String("username", player.Username))
		}
	} else {
		pos := &data.Position{}
		if err := pos.Decode(encoded); err != nil {
			log.Error("Failed to decode player position", zap.String("username", player.Username), zap.Error(err))
			return nil, errors.EDatabase.NewError(err)
		}

		log.Debug("Adding position to user", zap.String("username", player.Username), zap.String("position", encoded), zap.String("location", pos.Location))
		result.Position = pos
	}

	return result, nil
}

// ListPlayers - retrieve all users
func ListPlayers() ([]*data.Player, error) {
	pdb, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	rdb, err := connect.Redis()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var players []*Player
	if err := pdb.Find(&players).Error; err != nil {
		log.Error("Error fetching all users", zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	response := make([]*data.Player, len(players))
	for i, player := range players {
		response[i] = player.ToPlayer()

		key := fmt.Sprintf("%s:position", player.Username)
		encoded, err := rdb.Get(key).Result()
		if err != nil {
			if err != redis.Nil {
				log.Error("Failed to get position for player", zap.String("username", player.Username), zap.Error(err))
				return nil, errors.EDatabase.NewError(err)
			} else {
				log.Debug("No position for user", zap.String("username", player.Username))
			}
		} else {
			pos := &data.Position{}
			if err := pos.Decode(encoded); err != nil {
				log.Error("Failed to decode player position", zap.String("username", player.Username), zap.Error(err))
				return nil, errors.EDatabase.NewError(err)
			}

			log.Debug("Adding position to user", zap.String("username", player.Username), zap.String("position", encoded), zap.String("location", pos.Location))
			response[i].Position = pos
		}
	}

	return response, nil
}

// UpdatePlayer - update a player's username and/or password
func UpdatePlayer(id string, player *data.Player) (*data.Player, error) {
	db, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	updates := map[string]interface{}{
		"username":   player.Username,
		"updated_at": time.Now(),
	}

	var pw string
	if player.Password != nil {
		pw = *player.Password
	}

	if len(pw) != 0 {
		hashed, err := hashPassword(pw)
		if err != nil {
			return nil, errors.EInternal.NewError(err)
		}

		updates["password"] = hashed
	}

	var updated Player
	q := db.Where(&Player{Username: id}).Model(&updated).Updates(updates)
	if err := q.Error; err != nil {
		log.Error("Failed to patch player", zap.String("username", id), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	return updated.ToPlayer(), nil
}

// DeletePlayer - delete an existing user
func DeletePlayer(id string) error {
	pdb, err := connect.Postgres()
	if err != nil {
		return errors.EDatabaseConnection.NewError(err)
	}

	q := pdb.Delete(&Player{Username: id})
	if err := q.Error; err != nil {
		log.Error("Error deleting player", zap.String("username", id), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	if q.RowsAffected == 0 {
		log.Error("Cannot delete nonexistent player", zap.String("username", id), zap.Error(err))
		return errors.EDatabase.NewError("player does not exist")
	}

	rdb, err := connect.Redis()
	if err != nil {
		return errors.EDatabaseConnection.NewError(err)
	}

	pos := &data.Position{}
	key := fmt.Sprintf("%s:position", id)
	loc, err := rdb.Get(key).Result()
	if err != nil {
		if err != redis.Nil {
			log.Error("Error fetching location for user", zap.String("username", id), zap.Error(err))
			return errors.EDatabaseConnection.NewError(err)
		}
	} else {
		if err := pos.Decode(loc); err != nil {
			log.Error("Error decoding location record for user", zap.String("username", id), zap.String("data", loc), zap.Error(err))
			return err
		}

		if _, err := rdb.Del(key).Result(); err != nil {
			log.Error("Error deleting location record for user", zap.String("username", id), zap.Error(err))
		}

		key = fmt.Sprintf("location:%s", pos.Location)
		if _, err := rdb.SRem(key, id).Result(); err != nil {
			log.Error("Error removing deleted user from location", zap.String("username", id), zap.String("location", pos.Location), zap.Error(err))
			return errors.EDatabase.NewError(err)
		}
	}

	return nil
}

const saltLength int = 64

func hashPassword(password string) (string, error) {
	bs := make([]byte, saltLength)
	if _, err := rand.Read(bs); err != nil {
		log.Error("Error generating salt for new password", zap.Error(err))
		return "", errors.EInternal.NewError(nil)
	}

	salt := hex.EncodeToString(bs)
	salted := salt + password
	digest := sha256.Sum256([]byte(salted))
	return salt + hex.EncodeToString(digest[:]), nil
}

func testPassword(cleartext string, hashed string) bool {
	salt := hashed[:hex.EncodedLen(saltLength)]
	salted := salt + cleartext
	digest := sha256.Sum256([]byte(salted))
	return salt+hex.EncodeToString(digest[:]) == hashed
}

func init() {
	models = append(models, &Player{})
}
