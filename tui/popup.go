package tui

import "strings"

const popupInnerWidth = 40
const popupBorder = "+------------------------------------------+"

func popupCenter(rendered, plain string) string {
	if len(plain) >= popupInnerWidth {
		return plain[:popupInnerWidth]
	}
	left := (popupInnerWidth - len(plain)) / 2
	right := popupInnerWidth - len(plain) - left
	return strings.Repeat(" ", left) + rendered + strings.Repeat(" ", right)
}

func popupPadRight(rendered, plain string) string {
	if len(plain) >= popupInnerWidth {
		return plain[:popupInnerWidth]
	}
	return rendered + strings.Repeat(" ", popupInnerWidth-len(plain))
}

func popupFrame(lines []string) string {
	var b strings.Builder
	b.WriteString("\n" + popupBorder + "\n")
	for _, line := range lines {
		b.WriteString("| " + line + " |\n")
	}
	b.WriteString(popupBorder)
	return b.String()
}
