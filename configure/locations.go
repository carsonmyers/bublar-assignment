package configure

import "fmt"

// LocationsConfig - configuration struct for locations service
type LocationsConfig struct {
	Host     string
	Port     uint
	Protocol string
}

func (c *LocationsConfig) String() string {
	return fmt.Sprintf("%s://%s:%d", c.Protocol, c.Host, c.Port)
}

// DefaultLocationsConfig - configuration defaults which are overridden by options
var DefaultLocationsConfig = LocationsConfig{
	Host:     "0.0.0.0",
	Port:     49800,
	Protocol: "tcp",
}

var locationsConfig *LocationsConfig

// Locations - set the config
func Locations(config *LocationsConfig) {
	locationsConfig = config
}

// GetLocations - get the config
func GetLocations() *LocationsConfig {
	if locationsConfig == nil {
		return &DefaultLocationsConfig
	}

	return locationsConfig
}
