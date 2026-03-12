package config

import "fmt"

func defaultDevices(grid [][]int) map[string]string {
	devices := map[string]string{}
	for _, row := range grid {
		for _, slot := range row {
			key := fmt.Sprintf("%d", slot)
			if _, ok := devices[key]; ok {
				continue
			}
			devices[key] = fmt.Sprintf("/dev/disk%d", slot)
		}
	}
	return devices
}

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
				Devices: defaultDevices([][]int{{1}, {2}}),
			},
			{
				Name:    "Standard 4-bay",
				Rows:    2,
				Cols:    2,
				Grid:    [][]int{{1, 2}, {3, 4}},
				Devices: defaultDevices([][]int{{1, 2}, {3, 4}}),
			},
			{
				Name:    "Standard 6-bay",
				Rows:    3,
				Cols:    2,
				Grid:    [][]int{{1, 4}, {2, 5}, {3, 6}},
				Devices: defaultDevices([][]int{{1, 4}, {2, 5}, {3, 6}}),
			},
			{
				Name:    "Standard 12-bay",
				Rows:    3,
				Cols:    4,
				Grid:    [][]int{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}},
				Devices: defaultDevices([][]int{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}}),
			},
		},
	}
}
