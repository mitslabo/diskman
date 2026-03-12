package tui

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiRev    = "\x1b[7m"
	ansiCyan   = "\x1b[36m"
	ansiYellow = "\x1b[33m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
)

func style(s, code string) string {
	return code + s + ansiReset
}
