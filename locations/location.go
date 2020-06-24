package locations

import (
	"fmt"
	"time"

	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

var exp = 48 * time.Hour

// Location - a location within the game world
type Location struct {
	Name      string    `json:"name" gorm:"primary_key"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:timestamp"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"type:timestamp"`
}

// ToLocation - convert to universal data format
func (l *Location) ToLocation() *data.Location {
	return &data.Location{
		Name: l.Name,
		X:    l.X,
		Y:    l.Y,
	}
}

// CreateLocation - create a new location in the game world
func CreateLocation(location *data.Location) (*data.Location, error) {
	db, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var count uint64
	q := db.Model(&Location{}).Where(&Location{Name: location.Name}).Count(&count)
	if err := q.Error; err != nil {
		log.Error("Error counting existing locations", zap.Error(err))
		return nil, errors.EDatabase.NewError(location.Name)
	}

	if count != 0 {
		log.Error("Attempt to create a duplicate location", zap.String("name", location.Name), zap.Error(err))
		return nil, errors.EDuplicateLocation.NewError(err)
	}

	locationModel := Location{
		Name:      location.Name,
		X:         location.X,
		Y:         location.Y,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&locationModel).Error; err != nil {
		log.Error("Failed to store new location", zap.String("name", location.Name), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	return locationModel.ToLocation(), nil
}

// GetLocation - get a location by name
func GetLocation(id string) (*data.Location, error) {
	db, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var location Location
	if err := db.Where(&Location{Name: id}).First(&location).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.ENotFound.NewError(err)
		}

		log.Error("Error fetching location data", zap.String("name", location.Name), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	return location.ToLocation(), nil
}

// ListLocations - list all locations in the game
func ListLocations() ([]*data.Location, error) {
	db, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var locations []*Location
	if err := db.Find(&locations).Error; err != nil {
		log.Error("Error fetching all locations", zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	response := make([]*data.Location, len(locations))
	for i, location := range locations {
		response[i] = location.ToLocation()
	}

	return response, nil
}

// UpdateLocation - update the details of a location
func UpdateLocation(id string, location *data.Location) (*data.Location, error) {
	pdb, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var updated Location
	q := pdb.Raw(`
	UPDATE "location"
	SET
		"name" = ?,
		"x" = ?,
		"y" = ?,
		"updated_at" = ?
	WHERE
		"name" = ?
	RETURNING
		"location".*
	`, location.Name, location.X, location.Y, time.Now(), id).Scan(&updated)
	if err := q.Error; err != nil {
		log.Error("Failed to patch location", zap.String("name", id), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	if q.RowsAffected == 0 {
		return nil, errors.EDatabase.NewError("location not found")
	}

	rdb, err := connect.Redis()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	key := fmt.Sprintf("location:%s", id)
	members, err := rdb.SMembers(key).Result()
	if err != nil {
		log.Error("Error fetching players from updated location", zap.String("location", id), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	newKey := fmt.Sprintf("location:%s", location.Name)
	for _, member := range members {
		p := &data.Player{}
		if err := p.Decode(member); err != nil {
			log.Error("Error decoding player position for location update", zap.String("location", location.Name), zap.String("data", member), zap.Error(err))
			continue
		}

		p.Position.Location = location.Name

		if _, err := rdb.SAdd(newKey, p.Encode()).Result(); err != nil {
			log.Error("Error moving player to updated location", zap.String("location", location.Name), zap.String("playerData", member), zap.String("playerData", member), zap.Error(err))
			return nil, errors.EDatabase.NewError(err)
		}

		key := fmt.Sprintf("%s:position", p.Username)
		if _, err := rdb.Set(key, p.Position.Encode(), exp).Result(); err != nil {
			log.Error("Error updating player into updated location", zap.String("location", location.Name), zap.String("username", p.Username), zap.Error(err))
			continue
		}
	}

	if _, err := rdb.Del(key).Result(); err != nil {
		log.Error("Error removing updated location", zap.String("location", id), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	return updated.ToLocation(), nil
}

// ListPlayers - list all player positions within a location
func ListPlayers(location string) ([]*data.Player, error) {
	pdb, err := connect.Postgres()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	var l Location
	if err := pdb.Where(&Location{Name: location}).First(&l).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			log.Error("Cannot list users from unknown location", zap.String("location", location), zap.Error(err))
			return nil, errors.ENotFound.NewError(err)
		}

		log.Error("Error finding location", zap.String("location", location), zap.Error(err))
		return nil, errors.EDatabase.NewError(err)
	}

	rdb, err := connect.Redis()
	if err != nil {
		return nil, errors.EDatabaseConnection.NewError(err)
	}

	key := fmt.Sprintf("location:%s", location)
	members, err := rdb.SMembers(key).Result()
	if err != nil {
		return nil, errors.EDatabase.NewError(err)
	}

	log.Debug("Queried players from location", zap.String("key", key), zap.Int("players", len(members)))

	results := make([]*data.Player, len(members))
	for i, member := range members {
		results[i] = &data.Player{}
		if err := results[i].Decode(member); err != nil {
			return nil, errors.EDatabase.NewErrorf("failed to decode position `%s`", member).Wrap(err)
		}
	}

	return results, nil
}

// DeleteLocation - delete a location
func DeleteLocation(name string) error {
	pdb, err := connect.Postgres()
	if err != nil {
		return errors.EDatabaseConnection.NewError(err)
	}

	q := pdb.Delete(&Location{Name: name})
	if err := q.Error; err != nil {
		log.Error("Error deleting location", zap.String("name", name), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	if q.RowsAffected == 0 {
		log.Error("Location does not exist", zap.String("name", name))
		return errors.EDatabase.NewError("player does not exist")
	}

	rdb, err := connect.Redis()
	if err != nil {
		return errors.EDatabaseConnection.NewError(err)
	}

	key := fmt.Sprintf("location:%s", name)
	members, err := rdb.SMembers(key).Result()
	if err != nil {
		log.Error("Error fetching members in deleted location", zap.String("name", name), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	for _, member := range members {
		p := &data.Player{}
		if err := p.Decode(member); err != nil {
			log.Error("Error decoding member of deleted location", zap.String("name", name), zap.String("data", member), zap.Error(err))
			continue
		}

		if _, err := rdb.Del(fmt.Sprintf("%s:position", p.Username)).Result(); err != nil {
			log.Error("Error deleting location record for user", zap.String("name", name), zap.String("username", p.Username), zap.Error(err))
			continue
		}
	}

	if _, err := rdb.Del(key).Result(); err != nil {
		log.Error("Error deleting member list for location", zap.String("name", name), zap.Error(err))
		return errors.EDatabase.NewError(err)
	}

	return nil
}

func init() {
	models = append(models, &Location{})
}
