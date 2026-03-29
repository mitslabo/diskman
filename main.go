package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"diskman/config"
	"diskman/tui"
)

func defaultConfigPath() string {
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "diskman", "config.json")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "diskman", "config.json")
	}
	return "config.json"
}

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	dryRun := flag.Bool("dry-run", false, "do not execute ddrescue; simulate progress")
	debug := flag.Bool("debug", false, "enable debug mode with mock /dev/diskN paths and dry-run")
	flag.Parse()

	if *debug {
		*dryRun = true
	}

	configSpecified := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			configSpecified = true
		}
	})

	cfg, loadedFromFile, err := config.Load(*configPath, configSpecified)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config load error:", err)
		os.Exit(1)
	}

	program, err := tui.New(cfg, *dryRun, *debug, loadedFromFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "init error:", err)
		os.Exit(1)
	}
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "runtime error:", err)
		os.Exit(1)
	}
}
