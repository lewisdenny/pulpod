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

func GetLogger(loggingConfig *config.LoggingConfig) error {
	if Log == nil {
		// Initialize the logger only once
		err := newCustomLogger(loggingConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewCustomLogger creates and initializes a new CustomLogger.
func newCustomLogger(loggingConfig *config.LoggingConfig) error {
	var (
		err    error
		logger *zap.Logger
	)

	if loggingConfig.DevMode {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		fmt.Println(err)
		return err
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Println(err)
		}
	}()

	Log = &CustomLogger{
		SugaredLogger: logger.Sugar(),
	}
	return nil
}
