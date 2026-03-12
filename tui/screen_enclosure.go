package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *modelState) updateEnclosure(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.cfg.Enclosures)-1 {
			m.cursor++
		}
	case "enter":
		m.selectedEnc = m.cursor
		m.row, m.col = 0, 0
		m.srcSlot, m.dstSlot = -1, -1
		m.selectedOp = opCopy
		m.actionCursor = 0
		m.confirmCode = ""
		m.confirmInput = ""
		m.screen = scrSrc
	}
	return m, nil
}

func (m *modelState) viewEnclosure() string {
	var b strings.Builder
	b.WriteString("Select enclosure (Enter)\n")
	for i, e := range m.cfg.Enclosures {
		mark := "  "
		if i == m.cursor {
			mark = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s (%dx%d)\n", mark, e.Name, e.Rows, e.Cols))
	}
	b.WriteString("\nq: quit")
	return b.String()
}
