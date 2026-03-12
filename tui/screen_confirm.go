package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *modelState) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.confirmInput != m.confirmCode {
			m.status = "confirmation code does not match"
			return m, nil
		}
		if m.selectedOp == opErase {
			m.startEraseJob()
		} else {
			m.startCopyJob()
		}
		m.resetToSourceSelection()
	case "esc":
		if m.selectedOp == opErase {
			m.screen = scrAction
		} else {
			m.screen = scrDst
		}
		m.confirmInput = ""
	case "backspace":
		if len(m.confirmInput) > 0 {
			m.confirmInput = m.confirmInput[:len(m.confirmInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			if ch >= '0' && ch <= '9' {
				if len(m.confirmInput) < len(m.confirmCode) {
					m.confirmInput += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m *modelState) viewConfirm() string {
	e := m.cfg.Enclosures[m.selectedEnc]
	src := m.devicePath(e, m.srcSlot)
	if m.confirmCode == "" {
		m.confirmCode = randomConfirmCode()
	}
	title := "Confirm copy"
	detail := fmt.Sprintf("Enclosure: %s\nSource slot: %d (%s)\nDest slot:   %d (%s)", e.Name, m.srcSlot, src, m.dstSlot, m.devicePath(e, m.dstSlot))
	if m.selectedOp == opErase {
		title = "Confirm erase"
		detail = fmt.Sprintf("Enclosure: %s\nTarget slot: %d (%s)", e.Name, m.srcSlot, src)
	}
	masked := strings.Repeat("*", len(m.confirmInput))
	return fmt.Sprintf("%s\n\n%s\n\nCode: %s\nInput: %s\n\nType the code and press Enter.\nBackspace: delete  Esc: back", title, detail, m.confirmCode, masked)
}
