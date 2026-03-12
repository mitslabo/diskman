package tui

import (
	"fmt"
	"strings"

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
	const popupInnerWidth = 40
	center := func(rendered, plain string) string {
		if len(plain) >= popupInnerWidth {
			return plain[:popupInnerWidth]
		}
		left := (popupInnerWidth - len(plain)) / 2
		right := popupInnerWidth - len(plain) - left
		return strings.Repeat(" ", left) + rendered + strings.Repeat(" ", right)
	}
	pad := func(rendered, plain string) string {
		if len(plain) >= popupInnerWidth {
			return plain[:popupInnerWidth]
		}
		return rendered + strings.Repeat(" ", popupInnerWidth-len(plain))
	}
	slot := fmt.Sprintf("Slot%02d", m.srcSlot)
	var b strings.Builder
	b.WriteString("\n+------------------------------------------+\n")
	b.WriteString(fmt.Sprintf("| %s |\n", center("Select operation", "Select operation")))
	b.WriteString(fmt.Sprintf("| %s |\n", center(slot, slot)))
	b.WriteString(fmt.Sprintf("| %s |\n", center("", "")))
	for i, label := range actionLabels {
		plain := "  " + label
		line := plain
		if i == m.actionCursor {
			plain = "> " + label
			line = style(plain, ansiBgWhite+ansiBlack)
		}
		b.WriteString(fmt.Sprintf("| %s |\n", pad(line, plain)))
	}
	b.WriteString("+------------------------------------------+")
	return b.String()
}
