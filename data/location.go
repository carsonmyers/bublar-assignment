package data

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/carsonmyers/bublar-assignment/errors"
	"go.uber.org/zap"
)

// Location - local area within game world relative to other locations
type Location struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// Encode - encode a location as a string
func (l *Location) Encode() string {
	return fmt.Sprintf("%s:%d:%d", l.Name, l.X, l.Y)
}

// Decode - decode a location from a string
func (l *Location) Decode(data string) error {
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		return errors.EInternal.NewErrorf("invalid location encoding \"%s\"", parts)
	}

	l.Name = parts[0]

	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.EInternal.NewErrorf("invalid x coordinate \"%s\"", parts[1])
	}

	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return errors.EInternal.NewErrorf("invalid y coordinate \"%s\"", parts[2])
	}

	l.X = x
	l.Y = y

	return nil
}

// Position - point within the game world relative to a location
type Position struct {
	Location string `json:"location"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

// Encode - encode a position as a string
func (p *Position) Encode() string {
	return fmt.Sprintf("%s:%d:%d", p.Location, p.X, p.Y)
}

// Decode - decode a position from a string
func (p *Position) Decode(data string) error {
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		return errors.EInternal.NewErrorf("invalid position encoding \"%s\"", data)
	}

	p.Location = parts[0]

	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.EInternal.NewErrorf("invalid x coordinate \"%s\"", parts[1]).Wrap(err)
	}

	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return errors.EInternal.NewErrorf("invalid y coordinate \"%s\"", parts[2]).Wrap(err)
	}

	p.X = x
	p.Y = y

	log.Debug("Decoded position", zap.String("location", p.Location), zap.Int("x", p.X), zap.Int("y", p.Y))

	return nil
}
