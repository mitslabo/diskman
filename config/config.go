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
	_ = path // expanded
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if !configSpecified {
				return defaultCfg, false, nil // paths already resolved in defaultConfig()
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
	if userCfg.LogFile != "" {
		userCfg.LogFile = expandHome(userCfg.LogFile)
	} else {
		userCfg.LogFile = defaultCfg.LogFile
	}
	if userCfg.MapDir != "" {
		userCfg.MapDir = expandHome(userCfg.MapDir)
	} else {
		userCfg.MapDir = defaultCfg.MapDir
	}
	return userCfg, true, nil
}
