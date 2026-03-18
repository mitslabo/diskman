package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *modelState) updateInfo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.screen = scrSrc
	}
	return m, nil
}

func (m *modelState) viewInfo() string {
	e := m.cfg.Enclosures[m.selectedEnc]
	src := m.devicePath(e, m.srcSlot)
	model, serial, size := getDiskInfo(src)
	if model == "" {
		model = "N/A"
	}
	if serial == "" {
		serial = "N/A"
	}
	if size == "" {
		size = "N/A"
	}

	var b strings.Builder
	b.WriteString("Disk information\n\n")
	b.WriteString(fmt.Sprintf("Slot: %d\n", m.srcSlot))
	b.WriteString(fmt.Sprintf("Device: %s\n", src))
	b.WriteString(fmt.Sprintf("Model: %s\n", model))
	b.WriteString(fmt.Sprintf("Serial: %s\n", serial))
	b.WriteString(fmt.Sprintf("Capacity: %s\n", size))
	b.WriteString("\nEnter/Esc: back")
	return b.String()
}

func getDiskInfo(devicePath string) (model string, serial string, size string) {
	// Resolve symlinks (e.g. /dev/disk1 -> /dev/sda)
	realPath, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		realPath = devicePath
	}
	base := filepath.Base(realPath)

	// Attempt to read from /sys/block/<device>/device/{model,serial}
	sysRoot := filepath.Join("/sys/block", base)

	model = readSysfsValue(filepath.Join(sysRoot, "device", "model"))
	serial = readSysfsValue(filepath.Join(sysRoot, "device", "serial"))

	// Size is in 512-byte sectors
	szStr := readSysfsValue(filepath.Join(sysRoot, "size"))
	if szStr != "" {
		sz, err := strconv.ParseInt(strings.TrimSpace(szStr), 10, 64)
		if err == nil {
			size = formatBytes(sz * 512)
		}
	}

	return
}

func readSysfsValue(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func formatBytes(b int64) string {
	const (
		KB = 1 << (10 * 1)
		MB = 1 << (10 * 2)
		GB = 1 << (10 * 3)
		TB = 1 << (10 * 4)
	)
	f := float64(b)
	switch {
	case b >= TB:
		return fmt.Sprintf("%.2f TB", f/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.2f GB", f/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", f/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", f/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
