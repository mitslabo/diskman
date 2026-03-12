package config

func defaultConfig() Config {
	return Config{
		LogFile: "~/.local/share/diskman/jobs.jsonl",
		MapDir:  "~/.local/share/diskman/maps",
		Enclosures: []Enclosure{
			{
				Name:    "Standard 2-bay",
				Rows:    2,
				Cols:    1,
				Grid:    [][]int{{1}, {2}},
				Devices: map[string]string{},
			},
			{
				Name:    "Standard 4-bay",
				Rows:    2,
				Cols:    2,
				Grid:    [][]int{{1, 2}, {3, 4}},
				Devices: map[string]string{},
			},
			{
				Name:    "Standard 6-bay",
				Rows:    3,
				Cols:    2,
				Grid:    [][]int{{1, 4}, {2, 5}, {3, 6}},
				Devices: map[string]string{},
			},
			{
				Name:    "Standard 12-bay",
				Rows:    3,
				Cols:    4,
				Grid:    [][]int{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}},
				Devices: map[string]string{},
			},
		},
	}
}
