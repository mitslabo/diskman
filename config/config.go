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

// Load returns merged config from built-in defaults and optional user file.
func Load(path string) (Config, error) {
	cfg := defaultConfig()
	if path == "" {
		return cfg, nil
	}
	path = expandHome(path)
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, err
	}
	var userCfg Config
	if err := json.Unmarshal(b, &userCfg); err != nil {
		return Config{}, err
	}

	if userCfg.LogFile != "" {
		cfg.LogFile = userCfg.LogFile
	}
	if userCfg.MapDir != "" {
		cfg.MapDir = userCfg.MapDir
	}

	// Merge enclosures by name. Unknown name with full shape info is appended.
	for _, ue := range userCfg.Enclosures {
		matched := false
		for i := range cfg.Enclosures {
			if cfg.Enclosures[i].Name == ue.Name {
				matched = true
				if ue.Rows > 0 && ue.Cols > 0 && len(ue.Grid) > 0 {
					cfg.Enclosures[i].Rows = ue.Rows
					cfg.Enclosures[i].Cols = ue.Cols
					cfg.Enclosures[i].Grid = ue.Grid
				}
				if ue.Devices != nil {
					cfg.Enclosures[i].Devices = ue.Devices
				}
				break
			}
		}
		if !matched {
			if err := ue.Validate(); err != nil {
				return Config{}, err
			}
			cfg.Enclosures = append(cfg.Enclosures, ue)
		}
	}

	for _, e := range cfg.Enclosures {
		if err := e.Validate(); err != nil {
			return Config{}, err
		}
	}
	cfg.LogFile = expandHome(cfg.LogFile)
	cfg.MapDir = expandHome(cfg.MapDir)
	return cfg, nil
}
