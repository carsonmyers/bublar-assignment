package configure

import (
	"fmt"
	"net/url"
	"strings"
)

// PostgresConfig - configuration struct for postgres connections
type PostgresConfig struct {
	Host     string
	Port     uint
	Username string
	Password string `json:"-"`
	Database string
	SSLMode  string
}

// DefaultPostgresConfig - configuration defaults which are overridden by options
var DefaultPostgresConfig = PostgresConfig{
	Host:     "127.0.0.1",
	Port:     5432,
	Username: "bublar",
	Password: "bublar",
	Database: "bublar",
	SSLMode:  "disable",
}

func (c *PostgresConfig) String() string {
	var user *url.Userinfo
	if len(c.Password) > 0 {
		user = url.UserPassword(c.Username, c.Password)
	} else {
		user = url.User(c.Username)
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     c.Host,
		Path:     c.Database,
		RawQuery: fmt.Sprintf("sslmod=%s", c.SSLMode),
	}

	return u.String()
}

// ConnectionString get a gorm connection string for the config
func (c *PostgresConfig) ConnectionString() string {
	var b strings.Builder
	addParam(&b, "host", c.Host)
	addParam(&b, "port", c.Port)
	addParam(&b, "user", c.Username)
	addParam(&b, "password", c.Password)
	addParam(&b, "dbname", c.Database)
	addParam(&b, "sslmode", c.SSLMode)

	return b.String()
}

var postgresConfig *PostgresConfig

// Postgres set the configuration
func Postgres(config *PostgresConfig) {
	postgresConfig = config
}

// GetPostgres get the configuration
func GetPostgres() *PostgresConfig {
	if postgresConfig == nil {
		return &DefaultPostgresConfig
	}

	return postgresConfig
}
