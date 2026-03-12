package tui

import (
	"fmt"
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
	var b strings.Builder
	b.WriteString("Disk information (dummy)\n\n")
	b.WriteString(fmt.Sprintf("Slot: %d\n", m.srcSlot))
	b.WriteString(fmt.Sprintf("Device: %s\n", src))
	b.WriteString("Model: DUMMY-STORAGE-01\n")
	b.WriteString("Serial: DUMMY123456\n")
	b.WriteString("Capacity: 8 TB\n")
	b.WriteString("Health: GOOD\n")
	b.WriteString("\nEnter/Esc: back")
	return b.String()
}
