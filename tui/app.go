package tui

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"diskman/config"
	"diskman/logger"
	"diskman/model"
	"diskman/runner"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	scrEnclosure screen = iota
	scrSrc
	scrAction
	scrDst
	scrConfirm
	scrInfo
)

type operation int

const (
	opCopy operation = iota
	opErase
	opInfo
)

type modelState struct {
	cfg           config.Config
	dryRun        bool
	debug         bool
	screen        screen
	cursor        int
	row           int
	col           int
	selectedEnc   int
	srcSlot       int
	dstSlot       int
	selectedOp    operation
	actionCursor  int
	confirmCode   string
	confirmInput  string
	jobs          map[string]*model.Job
	jobOrder      []string
	jobCancels    map[string]context.CancelFunc
	updates       chan runner.Update
	log           *logger.JSONL
	status        string
	jobSel        int
	jobFocus      bool
	skipEnclosure bool
	cancelPopup   bool
	cancelChoice  int
	cancelJobID   string
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type updateMsg runner.Update

func waitUpdate(ch <-chan runner.Update) tea.Cmd {
	return func() tea.Msg { return updateMsg(<-ch) }
}

func defaultEnclosureIndex(enclosures []config.Enclosure) int {
	best := 0
	bestSlots := -1
	for i, e := range enclosures {
		slots := e.Rows * e.Cols
		if slots > bestSlots {
			best = i
			bestSlots = slots
		}
	}
	return best
}

func New(cfg config.Config, dryRun bool, debug bool, skipEnclosure bool) (*tea.Program, error) {
	if len(cfg.Enclosures) == 0 {
		return nil, fmt.Errorf("at least one enclosure is required")
	}
	l, err := logger.New(cfg.LogFile)
	if err != nil {
		return nil, err
	}
	m := &modelState{
		cfg:           cfg,
		dryRun:        dryRun,
		debug:         debug,
		screen:        scrEnclosure,
		selectedEnc:   defaultEnclosureIndex(cfg.Enclosures),
		skipEnclosure: skipEnclosure,
		srcSlot:       -1,
		dstSlot:       -1,
		selectedOp:    opCopy,
		jobs:          map[string]*model.Job{},
		jobCancels:    map[string]context.CancelFunc{},
		updates:       make(chan runner.Update, 256),
		log:           l,
	}
	if skipEnclosure {
		m.screen = scrSrc
	}
	p := tea.NewProgram(m)
	return p, nil
}

func (m *modelState) Init() tea.Cmd {
	return tea.Batch(tickCmd(), waitUpdate(m.updates))
}

func (m *modelState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		if s == "q" {
			if m.hasRunningJobs() {
				m.status = "running jobs exist; cancel or wait"
				return m, nil
			}
			if m.log != nil {
				_ = m.log.Close()
			}
			return m, tea.Quit
		}
		if m.screen == scrEnclosure {
			return m.updateEnclosure(msg)
		}
		if m.screen == scrSrc || m.screen == scrDst {
			return m.updateDisk(msg)
		}
		if m.screen == scrAction {
			return m.updateAction(msg)
		}
		if m.screen == scrConfirm {
			return m.updateConfirm(msg)
		}
		if m.screen == scrInfo {
			return m.updateInfo(msg)
		}
	case updateMsg:
		u := runner.Update(msg)
		if j, ok := m.jobs[u.JobID]; ok {
			j.Progress = u.Progress
			j.State = u.State
			if u.Err != nil {
				j.ErrMsg = u.Err.Error()
				m.status = "job error: " + j.ErrMsg
			}
			if u.Completed || u.Cancelled || u.State == model.JobError {
				delete(m.jobCancels, u.JobID)
			}
		}
		return m, waitUpdate(m.updates)
	case tickMsg:
		return m, tickCmd()
	}
	return m, nil
}

func (m *modelState) View() string {
	header := "diskman"
	if m.dryRun {
		header += " " + style("[DRY-RUN]", ansiYellow)
	}
	if m.debug {
		header += " " + style("[DEBUG]", ansiCyan)
	}
	head := style(header, ansiBold)
	if m.status != "" {
		head += "\n" + style(m.status, ansiRed)
	}
	switch m.screen {
	case scrEnclosure:
		return head + "\n\n" + m.viewEnclosure()
	case scrSrc, scrDst:
		return head + "\n\n" + m.viewDisk()
	case scrAction:
		return head + "\n\n" + m.viewAction()
	case scrConfirm:
		return head + "\n\n" + m.viewConfirm()
	case scrInfo:
		return head + "\n\n" + m.viewInfo()
	}
	return head
}

func (m *modelState) hasRunningJobs() bool {
	for _, j := range m.jobs {
		if j.State == model.JobRunning || j.State == model.JobPending {
			return true
		}
	}
	return false
}

func (m *modelState) activeJobIDs() []string {
	ids := make([]string, 0, len(m.jobOrder))
	for _, id := range m.jobOrder {
		j := m.jobs[id]
		if j == nil {
			continue
		}
		if j.State != model.JobPending && j.State != model.JobRunning {
			continue
		}
		if _, ok := m.jobCancels[id]; !ok {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func (m *modelState) slotLabelForJobPath(job *model.Job, path string) string {
	if job != nil {
		for _, e := range m.cfg.Enclosures {
			if e.Name != job.Name {
				continue
			}
			for k, v := range e.Devices {
				if v != path {
					continue
				}
				n, err := strconv.Atoi(k)
				if err != nil {
					return "Slot" + k
				}
				return fmt.Sprintf("Slot%02d", n)
			}
		}
	}
	end := len(path)
	start := end
	for start > 0 {
		c := path[start-1]
		if c < '0' || c > '9' {
			break
		}
		start--
	}
	if start < end {
		n, err := strconv.Atoi(path[start:end])
		if err == nil {
			return fmt.Sprintf("Slot%02d", n)
		}
		return "Slot" + path[start:end]
	}
	return "Slot??"
}

func (m *modelState) startCopyJob() {
	e := m.cfg.Enclosures[m.selectedEnc]
	src := m.devicePath(e, m.srcSlot)
	dst := m.devicePath(e, m.dstSlot)
	if src == "" || dst == "" {
		m.status = "src/dst device path not configured"
		return
	}
	id := model.NewJobID()
	job := &model.Job{
		ID:        id,
		Op:        "copy",
		Name:      e.Name,
		Src:       src,
		Dst:       dst,
		MapFile:   filepath.Join(m.cfg.MapDir, id+".map"),
		State:     model.JobPending,
		CreatedAt: time.Now(),
	}
	m.launchJob(job, m.dryRun)
	m.status = "copy job started"
}

func (m *modelState) startEraseJob() {
	e := m.cfg.Enclosures[m.selectedEnc]
	target := m.devicePath(e, m.srcSlot)
	if target == "" {
		m.status = "target device path not configured"
		return
	}
	id := model.NewJobID()
	job := &model.Job{
		ID:        id,
		Op:        "erase",
		Name:      e.Name,
		Src:       target,
		Dst:       target,
		MapFile:   filepath.Join(m.cfg.MapDir, id+".map"),
		State:     model.JobPending,
		CreatedAt: time.Now(),
	}
	m.launchJob(job, m.dryRun)
	m.status = "erase job started"
}

func (m *modelState) launchJob(job *model.Job, dryRun bool) {
	m.jobs[job.ID] = job
	m.jobOrder = append(m.jobOrder, job.ID)
	ctx, cancel := context.WithCancel(context.Background())
	m.jobCancels[job.ID] = cancel
	_ = m.log.Log(map[string]any{"time": time.Now().Format(time.RFC3339), "event": "job_start", "id": job.ID, "op": job.Op, "src": job.Src, "dst": job.Dst, "dryRun": dryRun})
	runner.StartJob(ctx, *job, dryRun, m.updates)
}

func (m *modelState) devicePath(e config.Enclosure, slot int) string {
	if p := e.Devices[fmt.Sprintf("%d", slot)]; p != "" {
		return p
	}
	if m.debug {
		return fmt.Sprintf("/dev/disk%d", slot)
	}
	return ""
}

func (m *modelState) isDeviceBusy(path string) bool {
	if path == "" {
		return false
	}
	for _, j := range m.jobs {
		if j == nil {
			continue
		}
		if j.State != model.JobPending && j.State != model.JobRunning {
			continue
		}
		if j.Src == path || j.Dst == path {
			return true
		}
	}
	return false
}

func (m *modelState) startConfirmation() {
	m.confirmCode = randomConfirmCode()
	m.confirmInput = ""
	m.screen = scrConfirm
}

func (m *modelState) resetToSourceSelection() {
	m.screen = scrSrc
	m.row, m.col = 0, 0
	m.srcSlot, m.dstSlot = -1, -1
	m.selectedOp = opCopy
	m.actionCursor = 0
	m.jobFocus = false
	m.cancelPopup = false
	m.cancelChoice = 0
	m.cancelJobID = ""
	m.confirmCode = ""
	m.confirmInput = ""
}

func randomConfirmCode() string {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "0000"
	}
	v := int(binary.BigEndian.Uint16(b[:])) % 10000
	return fmt.Sprintf("%04d", v)
}

func (m *modelState) diskUsageLabel(path string) (string, string) {
	if path == "" {
		return "", ""
	}
	copyN := 0
	for _, id := range m.jobOrder {
		j := m.jobs[id]
		if j == nil {
			continue
		}
		if j.State != model.JobPending && j.State != model.JobRunning {
			continue
		}
		op := j.Op
		if op == "" {
			if j.Src == j.Dst {
				op = "erase"
			} else {
				op = "copy"
			}
		}
		switch op {
		case "erase":
			if j.Src == path || j.Dst == path {
				return "E", "削除中"
			}
		default:
			copyN++
			if j.Src == path {
				return fmt.Sprintf("S%d", copyN), fmt.Sprintf("JOB%d コピー元", copyN)
			}
			if j.Dst == path {
				return fmt.Sprintf("D%d", copyN), fmt.Sprintf("JOB%d コピー先", copyN)
			}
		}
	}
	return "", ""
}
