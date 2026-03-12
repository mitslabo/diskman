package main

import (
	"flag"
	"fmt"
	"os"

	"diskman/config"
	"diskman/tui"
)

func main() {
	configPath := flag.String("config", "~/.config/diskman/config.json", "path to config file")
	dryRun := flag.Bool("dry-run", false, "do not execute ddrescue; simulate progress")
	debug := flag.Bool("debug", false, "enable debug mode with mock /dev/diskN paths and dry-run")
	flag.Parse()

	if *debug {
		*dryRun = true
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config load error:", err)
		os.Exit(1)
	}

	program, err := tui.New(cfg, *dryRun, *debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, "init error:", err)
		os.Exit(1)
	}
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "runtime error:", err)
		os.Exit(1)
	}
}
