package tui

import (
	"fmt"
	"strings"
	"time"

	"diskman/model"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *modelState) updateDisk(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	e := m.cfg.Enclosures[m.selectedEnc]
	activeIDs := m.activeJobIDs()
	if len(activeIDs) == 0 {
		m.jobSel = 0
		m.jobFocus = false
	} else if m.jobSel >= len(activeIDs) {
		m.jobSel = len(activeIDs) - 1
	}
	switch msg.String() {
	case "up":
		if m.jobFocus {
			if m.jobSel > 0 {
				m.jobSel--
			} else {
				m.jobFocus = false
			}
		} else if m.row > 0 {
			m.row--
		}
	case "down":
		if m.jobFocus {
			if m.jobSel < len(activeIDs)-1 {
				m.jobSel++
			}
		} else if m.row < e.Rows-1 {
			m.row++
		} else if len(activeIDs) > 0 {
			m.jobFocus = true
		}
	case "k":
		if m.row > 0 {
			m.row--
		}
	case "j":
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
	case "c":
		if len(activeIDs) == 0 || !m.jobFocus {
			m.status = "no cancellable jobs"
			return m, nil
		}
		id := activeIDs[m.jobSel]
		cancel := m.jobCancels[id]
		cancel()
		if m.log != nil {
			_ = m.log.Log(map[string]any{"time": time.Now().Format(time.RFC3339), "event": "job_cancel_requested", "id": id})
		}
		m.status = "cancel requested"
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
		m.jobFocus = false
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
	b.WriteString("\nArrows: seamless disk/job move  h/j/k/l: move disk  Enter: select  Esc: back")
	b.WriteString("\n\nJob status\n")
	activeIDs := m.activeJobIDs()
	if len(activeIDs) == 0 {
		m.jobSel = 0
	} else if m.jobSel >= len(activeIDs) {
		m.jobSel = len(activeIDs) - 1
	}
	for i, id := range activeIDs {
		j := m.jobs[id]
		if j == nil {
			continue
		}
		state := string(j.State)
		if j.State == model.JobRunning {
			state = style(state, ansiGreen)
		}
		op := j.Op
		if op == "" {
			op = "copy"
		}
		mark := " "
		if i == m.jobSel {
			mark = ">"
		}
		line := fmt.Sprintf("%s %s op:%s pass:%d %5.1f%% rate:%s state:%s", mark, j.ID, op, j.Progress.Pass, j.Progress.Percent, j.Progress.Rate, state)
		if i == m.jobSel && m.jobFocus {
			line = style(line, ansiRev)
		}
		b.WriteString(line + "\n")
	}
	if len(activeIDs) == 0 {
		b.WriteString("- none\n")
	} else {
		b.WriteString("Down at last disk row: enter jobs  Up at first job: return to disks\n")
		b.WriteString("c: cancel selected job (job selection active)\n")
	}
	return b.String()
}
