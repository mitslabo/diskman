package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var actionLabels = []string{"コピー", "削除", "情報表示"}

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
	e := m.cfg.Enclosures[m.selectedEnc]
	src := m.devicePath(e, m.srcSlot)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Select operation for slot %d (%s)\n\n", m.srcSlot, src))
	for i, label := range actionLabels {
		mark := "  "
		if i == m.actionCursor {
			mark = "> "
		}
		b.WriteString(mark + label + "\n")
	}
	b.WriteString("\nEnter: select  Esc: back  Tab: jobs")
	return b.String()
}
