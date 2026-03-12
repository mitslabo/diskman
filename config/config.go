package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func expandHome(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// Load returns config and whether it was loaded from a file.
// If configSpecified is true, missing file is an error.
func Load(path string, configSpecified bool) (Config, bool, error) {
	defaultCfg := defaultConfig()
	if path == "" {
		return defaultCfg, false, nil
	}
	path = expandHome(path)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if !configSpecified {
				return defaultCfg, false, nil
			}
			return Config{}, false, err
		}
		return Config{}, false, err
	}
	var userCfg Config
	if err := json.Unmarshal(b, &userCfg); err != nil {
		return Config{}, false, err
	}
	for _, e := range userCfg.Enclosures {
		if err := e.Validate(); err != nil {
			return Config{}, false, err
		}
	}
	userCfg.LogFile = expandHome(userCfg.LogFile)
	userCfg.MapDir = expandHome(userCfg.MapDir)
	return userCfg, true, nil
}
