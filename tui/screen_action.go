package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

var actionLabels = []string{"COPY", "ERASE", "INFO"}

func (m *modelState) updateAction(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case "down", "j":
		if m.actionCursor < len(actionLabels)-1 {
			m.actionCursor++
		}
	case "enter":
		switch m.actionCursor {
		case 0:
			m.selectedOp = opCopy
			m.dstSlot = -1
			m.screen = scrDst
		case 1:
			m.selectedOp = opErase
			m.dstSlot = -1
			m.startConfirmation()
		case 2:
			m.selectedOp = opInfo
			m.screen = scrInfo
		}
	case "esc":
		m.screen = scrSrc
		m.srcSlot = -1
	}
	return m, nil
}

func (m *modelState) viewAction() string {
	slot := fmt.Sprintf("Slot%02d", m.srcSlot)
	lines := []string{
		popupCenter("Select operation", "Select operation"),
		popupCenter(slot, slot),
		popupCenter("", ""),
	}
	for i, label := range actionLabels {
		plain := "  " + label
		line := plain
		if i == m.actionCursor {
			plain = "> " + label
			line = style(plain, ansiBgWhite+ansiBlack)
		}
		lines = append(lines, popupPadRight(line, plain))
	}
	return popupFrame(lines)
}
