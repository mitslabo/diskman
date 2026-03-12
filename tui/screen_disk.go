package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *modelState) updateDisk(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	e := m.cfg.Enclosures[m.selectedEnc]
	switch msg.String() {
	case "up", "k":
		if m.row > 0 {
			m.row--
		}
	case "down", "j":
		if m.row < e.Rows-1 {
			m.row++
		}
	case "left", "h":
		if m.col > 0 {
			m.col--
		}
	case "right", "l":
		if m.col < e.Cols-1 {
			m.col++
		}
	case "enter":
		slot := e.Grid[m.row][m.col]
		path := m.devicePath(e, slot)
		if m.isDeviceBusy(path) {
			m.status = "selected disk is in use by a running job"
			return m, nil
		}
		if m.screen == scrSrc {
			m.srcSlot = slot
			m.actionCursor = 0
			m.selectedOp = opCopy
			m.screen = scrAction
		} else {
			if slot == m.srcSlot {
				m.status = "src and dst must differ"
				return m, nil
			}
			m.dstSlot = slot
			m.startConfirmation()
		}
	case "esc":
		if m.screen == scrDst {
			m.screen = scrAction
			m.dstSlot = -1
		} else {
			m.screen = scrEnclosure
		}
	}
	return m, nil
}

func (m *modelState) viewDisk() string {
	e := m.cfg.Enclosures[m.selectedEnc]
	label := "Select source slot"
	if m.screen == scrDst {
		label = "Select destination slot"
	}
	var b strings.Builder
	b.WriteString(label + "\n\n")
	for r := 0; r < e.Rows; r++ {
		for c := 0; c < e.Cols; c++ {
			slot := e.Grid[r][c]
			status := " "
			if slot == m.dstSlot {
				status = "D"
			}
			name := m.devicePath(e, slot)
			usageLabel, _ := m.diskUsageLabel(name)
			if usageLabel == "" {
				usageLabel = " "
			}
			if slot == m.srcSlot {
				usageLabel = "S"
			}
			cell := fmt.Sprintf("[%2s] Slot%02d", usageLabel, slot)
			if status != " " {
				cell += " " + style("("+status+")", ansiBold)
			}
			if slot == m.srcSlot {
				cell = style(cell, ansiCyan)
			}
			if slot == m.dstSlot {
				cell = style(cell, ansiYellow)
			}
			if r == m.row && c == m.col {
				cell = style(cell, ansiRev)
			}
			b.WriteString(cell)
			if c < e.Cols-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n")
	}
	if m.screen == scrSrc {
		b.WriteString("\nArrows: move  Enter: select  Esc: back  Tab: jobs")
	} else {
		b.WriteString("\nArrows: move  Enter: select  Esc: back")
	}
	return b.String()
}
