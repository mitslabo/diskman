package tui

import (
	"fmt"
	"strings"
	"time"

	"diskman/model"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *modelState) updateJobs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.jobSel > 0 {
			m.jobSel--
		}
	case "down", "j":
		if m.jobSel < len(m.jobOrder)-1 {
			m.jobSel++
		}
	case "c":
		if len(m.jobOrder) == 0 {
			return m, nil
		}
		id := m.jobOrder[m.jobSel]
		if cancel, ok := m.jobCancels[id]; ok {
			cancel()
			_ = m.log.Log(map[string]any{"time": time.Now().Format(time.RFC3339), "event": "job_cancel_requested", "id": id})
			m.status = "cancel requested"
		}
	}
	return m, nil
}

func (m *modelState) viewJobs() string {
	var b strings.Builder
	b.WriteString("Jobs\n\n")
	if len(m.jobOrder) == 0 {
		b.WriteString("(no jobs)\n")
		b.WriteString("\nTab: back")
		return b.String()
	}
	for i, id := range m.jobOrder {
		j := m.jobs[id]
		if j == nil {
			continue
		}
		mark := " "
		if i == m.jobSel {
			mark = ">"
		}
		state := string(j.State)
		op := j.Op
		if op == "" {
			op = "copy"
		}
		if j.State == model.JobRunning {
			state = style(state, ansiGreen)
		}
		if j.State == model.JobError || j.State == model.JobCancelled {
			state = style(state, ansiRed)
		}
		line := fmt.Sprintf("%s %s op:%s %s->%s pass:%d %5.1f%% rate:%s state:%s", mark, j.ID, op, j.Src, j.Dst, j.Progress.Pass, j.Progress.Percent, j.Progress.Rate, state)
		if m.dryRun {
			line += " " + style("[DRY-RUN]", ansiYellow)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\nUp/Down: select  c: cancel  Tab: back")
	return b.String()
}
