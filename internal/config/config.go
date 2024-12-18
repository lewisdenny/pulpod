package config

import (
	"errors"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type ContainerManagerConfig struct {
	Socket string `koanf:"socket"`
	Flavor string `koanf:"flavor"`
}

type LoggingConfig struct {
	DevMode bool `koanf:"devmode"`
}

type Config struct {
	ContainerManagerConfig ContainerManagerConfig `koanf:"containermanager"`
	LoggingConfig          LoggingConfig          `koanf:"logging"`
}

func Configure(path string) (*Config, error) {
	config := &Config{}
	err := loadConfig(config, path)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// loadConfig function  
// Load config first from file, then from environment vars.
// Environment vars take precendent.
func loadConfig(config *Config, path string) error {
	koanfInstance := koanf.New(".")

	err := koanfInstance.Load(file.Provider("/etc/pulpod/config.toml"), toml.Parser())
	if errRel := koanfInstance.Load(file.Provider(path), toml.Parser()); errRel != nil && err != nil {
		return errors.New("unable to load service configuration from known locations")
	}

	// Example env var:
	// PULPOD_CONTAINERMANAGER_FLAVOR=podman
	// PULPOD_ will be trimmed and _ replaced with .
	// resulting in containermanager.flavor: podman
	koanfInstance.Load(env.Provider("PULPOD_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "PULPOD_")), "_", ".", -1)
	}), nil)

	return koanfInstance.Unmarshal("", config)
}
