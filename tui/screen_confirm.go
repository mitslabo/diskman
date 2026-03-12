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
	const popupInnerWidth = 40
	center := func(rendered, plain string) string {
		if len(plain) >= popupInnerWidth {
			return plain[:popupInnerWidth]
		}
		left := (popupInnerWidth - len(plain)) / 2
		right := popupInnerWidth - len(plain) - left
		return strings.Repeat(" ", left) + rendered + strings.Repeat(" ", right)
	}
	if m.confirmCode == "" {
		m.confirmCode = randomConfirmCode()
	}
	title := "Confirm COPY"
	actionLine := fmt.Sprintf("COPY Slot%02d -> Slot%02d", m.srcSlot, m.dstSlot)
	if m.selectedOp == opErase {
		title = "Confirm ERASE"
		actionLine = fmt.Sprintf("ERASE Slot%02d", m.srcSlot)
	}
	masked := strings.Repeat("*", len(m.confirmInput))
	var b strings.Builder
	b.WriteString("\n+------------------------------------------+\n")
	b.WriteString(fmt.Sprintf("| %s |\n", center(title, title)))
	b.WriteString(fmt.Sprintf("| %s |\n", center(actionLine, actionLine)))
	b.WriteString(fmt.Sprintf("| %s |\n", center("", "")))
	b.WriteString(fmt.Sprintf("| %s |\n", center("Code: "+m.confirmCode, "Code: "+m.confirmCode)))
	b.WriteString(fmt.Sprintf("| %s |\n", center("Input: "+masked, "Input: "+masked)))
	b.WriteString("+------------------------------------------+")
	return b.String()
}
