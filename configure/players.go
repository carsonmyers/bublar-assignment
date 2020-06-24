package configure

import "fmt"

// PlayersConfig - configuration struct for players service
type PlayersConfig struct {
	Host     string
	Port     uint
	Protocol string
}

func (c *PlayersConfig) String() string {
	return fmt.Sprintf("%s://%s:%d", c.Protocol, c.Host, c.Port)
}

// DefaultPlayersConfig - configuration defaults which are overridden by options
var DefaultPlayersConfig = PlayersConfig{
	Host:     "0.0.0.0",
	Port:     49801,
	Protocol: "tcp",
}

var playersConfig *PlayersConfig

// Players - set the config
func Players(config *PlayersConfig) {
	playersConfig = config
}

// GetPlayers - get the config
func GetPlayers() *PlayersConfig {
	if playersConfig == nil {
		return &DefaultPlayersConfig
	}

	return playersConfig
}
