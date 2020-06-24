package logger

import (
	"log"

	"github.com/carsonmyers/bublar-assignment/configure"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// GetLogger get the global logger
func GetLogger() *zap.Logger {
	return logger
}

// New create a new instance of a logger
func New() *zap.Logger {
	loggerConfig := configure.GetLogger()
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = zap.NewAtomicLevelAt(loggerConfig.Level)
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Could not initialize logging utility: %s", err)
	}

	defer logger.Sync()
	return logger
}

func init() {
	loggerConfig := &configure.DefaultLoggerConfig
	envconfig.MustProcess("log", loggerConfig)
	configure.Logger(loggerConfig)

	logger = New()
}
