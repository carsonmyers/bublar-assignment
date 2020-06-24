package data

import (
	"fmt"
	"strings"

	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/carsonmyers/bublar-assignment/logger"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// Player - player within the game world
type Player struct {
	Username string    `json:"username"`
	Password *string   `json:"password,omitempty"`
	Position *Position `json:"position"`
}

// Encode - encode a user and their position as a string
func (p *Player) Encode() string {
	if p.Position != nil {
		return fmt.Sprintf("%s:%s", p.Username, p.Position.Encode())
	}

	return fmt.Sprintf("%s:%s", p.Username, (&Position{}).Encode())
}

// Decode - decode a user from a string
func (p *Player) Decode(data string) error {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		return errors.EInternal.NewErrorf("invalid user encoding \"%s\"", data)
	}

	p.Username = parts[0]
	p.Position = &Position{}
	if err := p.Position.Decode(parts[1]); err != nil {
		return err
	}

	log.Debug("Decoded player", zap.String("username", p.Username), zap.String("position", fmt.Sprintf("%v", *p.Position)))

	return nil
}
