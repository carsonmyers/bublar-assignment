package configure

import "fmt"

// APIConfig - configuration struct for API service
type APIConfig struct {
	Host        string
	Port        uint
	Protocol    string
	BasePath    string
	Name        string
	Session     string
	EnableAdmin bool
}

func (c *APIConfig) String() string {
	return fmt.Sprintf("%s://%s:%d", c.Protocol, c.Host, c.Port)
}

// DefaultAPIConfig - configuration defaults which are overridden by options
var DefaultAPIConfig = APIConfig{
	Host:        "0.0.0.0",
	Port:        62880,
	Protocol:    "http",
	BasePath:    "/v1",
	Name:        "API",
	EnableAdmin: false,
}

var apiConfig *APIConfig

// API - set the config
func API(config *APIConfig) {
	apiConfig = config
}

// GetAPI - get the config
func GetAPI() *APIConfig {
	if apiConfig == nil {
		return &DefaultAPIConfig
	}

	return apiConfig
}
