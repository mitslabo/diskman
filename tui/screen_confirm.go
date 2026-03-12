package tui

import (
	"fmt"

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
	if m.confirmCode == "" {
		m.confirmCode = randomConfirmCode()
	}
	title := "Confirm COPY"
	actionLine := fmt.Sprintf("COPY Slot%02d -> Slot%02d", m.srcSlot, m.dstSlot)
	if m.selectedOp == opErase {
		title = "Confirm ERASE"
		actionLine = fmt.Sprintf("ERASE Slot%02d", m.srcSlot)
	}
	lines := []string{
		popupCenter(title, title),
		popupCenter(actionLine, actionLine),
		popupCenter("", ""),
		popupCenter("Code: "+m.confirmCode, "Code: "+m.confirmCode),
		popupCenter("Input: "+m.confirmInput, "Input: "+m.confirmInput),
	}
	return popupFrame(lines)
}
