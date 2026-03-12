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
		m.cancelPopup = false
		m.cancelJobID = ""
	} else if m.jobSel >= len(activeIDs) {
		m.jobSel = len(activeIDs) - 1
	}
	if m.cancelPopup {
		switch msg.String() {
		case "left", "h":
			m.cancelChoice = 0
		case "right", "l":
			m.cancelChoice = 1
		case "esc":
			m.cancelPopup = false
			m.cancelJobID = ""
		case "enter":
			if m.cancelChoice == 0 {
				if cancel, ok := m.jobCancels[m.cancelJobID]; ok {
					cancel()
					if m.log != nil {
						_ = m.log.Log(map[string]any{"time": time.Now().Format(time.RFC3339), "event": "job_cancel_requested", "id": m.cancelJobID})
					}
					m.status = "cancel requested"
				}
			}
			m.cancelPopup = false
			m.cancelJobID = ""
		}
		return m, nil
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
	case "enter":
		if len(activeIDs) > 0 && m.jobFocus {
			m.cancelPopup = true
			m.cancelChoice = 0
			m.cancelJobID = activeIDs[m.jobSel]
			return m, nil
		}
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
		m.cancelPopup = false
		m.cancelJobID = ""
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
	}
	if m.cancelPopup {
		job := m.jobs[m.cancelJobID]
		actionLine := "COPY Slot?? -> Slot??"
		if job != nil {
			src := m.slotLabelForJobPath(job, job.Src)
			dst := m.slotLabelForJobPath(job, job.Dst)
			op := job.Op
			if op == "" {
				if job.Src == job.Dst {
					op = "erase"
				} else {
					op = "copy"
				}
			}
			if op == "erase" {
				actionLine = fmt.Sprintf("ERASE %s", src)
			} else {
				actionLine = fmt.Sprintf("COPY %s -> %s", src, dst)
			}
		}
		yes := "[YES]"
		no := "[no]"
		if m.cancelChoice == 0 {
			yes = "[YES]"
			no = "[no]"
		} else {
			yes = "[yes]"
			no = "[NO]"
		}
		choiceLine := fmt.Sprintf("%s   %s", yes, no)
		if m.cancelChoice == 0 {
			yes = style(yes, ansiBgWhite+ansiBlack)
		} else {
			no = style(no, ansiBgWhite+ansiBlack)
		}
		choiceLine = popupCenter(fmt.Sprintf("%s   %s", yes, no), fmt.Sprintf("%s   %s", "[YES]", "[NO]"))
		b.WriteString(popupFrame([]string{
			popupCenter("Cancel selected job?", "Cancel selected job?"),
			popupCenter(actionLine, actionLine),
			popupCenter("", ""),
			choiceLine,
		}) + "\n")
	}
	return b.String()
}
