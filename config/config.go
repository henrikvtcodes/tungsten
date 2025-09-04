package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pelletier/go-toml/v2"
	"io"
	"os"
	"path"

	"github.com/carlmjohnson/truthy"
)

var (
	DevMode = truthy.Value(os.Getenv("TUNGSTEN_DEV_MODE"))
)

type WrappedServerConfig struct {
	DNSConfig  *ServerConfigFile
	SocketPath string
	ConfigPath string
}

func LoadFromPath(fPath string) (*ServerConfigFile, error) {
	var config = new(ServerConfigFile)
	if err := config.InitializeAndSetDefaults(); err != nil {
		return nil, err
	}

	// Read file into byte slice
	file, fErr := os.Open(fPath)
	if fErr != nil {
		err := errors.Join(fmt.Errorf("failed to open config file"), fErr)
		return nil, err
	}
	fileBytes, readErr := io.ReadAll(file)
	if readErr != nil {
		err := errors.Join(fmt.Errorf("failed to read config file"), readErr)
		return nil, err
	}
	if err := file.Close(); err != nil {
		err = errors.Join(fmt.Errorf("failed to close config file"), err)
		return nil, err
	}

	// Multi format file reading!!
	switch path.Ext(fPath) {
	case "toml":
		if err := toml.Unmarshal(fileBytes, config); err != nil {
			err = errors.Join(fmt.Errorf("failed to unmarshal toml"), err)
			return nil, err
		}
	case "json":
		if err := json.Unmarshal(fileBytes, config); err != nil {
			err = errors.Join(fmt.Errorf("failed to unmarshal json"), err)
			return nil, err
		}
	case "yaml":
		if err := yaml.Unmarshal(fileBytes, config); err != nil {
			err = errors.Join(fmt.Errorf("failed to unmarshal yaml"), err)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", path.Ext(fPath))
	}

	return config, nil
}
