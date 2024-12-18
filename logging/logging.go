package logging

import (
	"fmt"

	"github.com/openstack-tooling/pulpod/internal/config"
	"go.uber.org/zap"
)

var (
	Log *CustomLogger
)

// CustomLogger wraps the standard library logger and adds more methods.
type CustomLogger struct {
	*zap.SugaredLogger
}

func GetLogger(LoggingConfig *config.LoggingConfig) error {
	if Log == nil {
		// Initialize the logger only once
		err := newCustomLogger(LoggingConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewCustomLogger creates and initializes a new CustomLogger.
func newCustomLogger(LoggingConfig *config.LoggingConfig) error {
	var (
		err    error
		logger *zap.Logger
	)

	if LoggingConfig.DevMode {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		fmt.Println(err)
		return err
	}

	defer logger.Sync()

	Log = &CustomLogger{
		SugaredLogger: logger.Sugar(),
	}
	return nil
}
