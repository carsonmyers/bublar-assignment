package configure

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// LoggerConfig ENV config for the logger
type LoggerConfig struct {
	Level zapcore.Level
}

func (c *LoggerConfig) String() string {
	return fmt.Sprintf("logging=%s", c.Level)
}

// DefaultLoggerConfig used if no config override is set
var DefaultLoggerConfig = LoggerConfig{
	Level: zapcore.ErrorLevel,
}

var loggerConfig *LoggerConfig

// Logger set the config
func Logger(config *LoggerConfig) {
	loggerConfig = config
}

// GetLogger get the config
func GetLogger() *LoggerConfig {
	if loggerConfig == nil {
		return &DefaultLoggerConfig
	}

	return loggerConfig
}
