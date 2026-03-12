package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiRev     = "\x1b[7m"
	ansiBlack   = "\x1b[30m"
	ansiBgWhite = "\x1b[47m"
	ansiCyan    = "\x1b[36m"
	ansiYellow  = "\x1b[33m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
)

func style(s, code string) string {
	st := lipgloss.NewStyle()
	if strings.Contains(code, ansiBold) {
		st = st.Bold(true)
	}
	if strings.Contains(code, ansiRev) {
		st = st.Reverse(true)
	}
	if strings.Contains(code, ansiBlack) {
		st = st.Foreground(lipgloss.Color("0"))
	}
	if strings.Contains(code, ansiBgWhite) {
		st = st.Background(lipgloss.Color("15"))
	}
	if strings.Contains(code, ansiCyan) {
		st = st.Foreground(lipgloss.Color("6"))
	}
	if strings.Contains(code, ansiYellow) {
		st = st.Foreground(lipgloss.Color("3"))
	}
	if strings.Contains(code, ansiRed) {
		st = st.Foreground(lipgloss.Color("1"))
	}
	if strings.Contains(code, ansiGreen) {
		st = st.Foreground(lipgloss.Color("2"))
	}
	return st.Render(s)
}
