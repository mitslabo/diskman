package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"diskman/model"
)

type Update struct {
	JobID     string
	Progress  model.Progress
	State     model.JobState
	Err       error
	Completed bool
	Cancelled bool
}

func StartJob(ctx context.Context, job model.Job, dryRun bool, out chan<- Update) {
	go func() {
		out <- Update{JobID: job.ID, State: model.JobRunning}

		if job.Op == "erase" {
			var err error
			if dryRun {
				err = runDryErase(ctx, job.ID, out)
			} else {
				err = runErase(ctx, job, out)
			}
			if err != nil {
				if ctx.Err() != nil {
					out <- Update{JobID: job.ID, State: model.JobCancelled, Cancelled: true}
					return
				}
				out <- Update{JobID: job.ID, State: model.JobError, Err: err}
				return
			}
			out <- Update{JobID: job.ID, State: model.JobDone, Completed: true}
			return
		}

		// copy: ddrescue 3-pass
		if err := os.MkdirAll(filepath.Dir(job.MapFile), 0o755); err != nil {
			out <- Update{JobID: job.ID, State: model.JobError, Err: err}
			return
		}
		for pass := 1; pass <= 3; pass++ {
			select {
			case <-ctx.Done():
				out <- Update{JobID: job.ID, State: model.JobCancelled, Cancelled: true}
				return
			default:
			}
			if dryRun {
				if err := runDryPass(ctx, job.ID, pass, out); err != nil {
					if ctx.Err() != nil {
						out <- Update{JobID: job.ID, State: model.JobCancelled, Cancelled: true}
						return
					}
					out <- Update{JobID: job.ID, State: model.JobError, Err: err}
					return
				}
			} else {
				if err := runRealPass(ctx, job, pass, out); err != nil {
					if ctx.Err() != nil {
						out <- Update{JobID: job.ID, State: model.JobCancelled, Cancelled: true}
						return
					}
					out <- Update{JobID: job.ID, State: model.JobError, Err: err}
					return
				}
			}
		}
		out <- Update{JobID: job.ID, State: model.JobDone, Completed: true}
	}()
}

func runDryPass(ctx context.Context, jobID string, pass int, out chan<- Update) error {
	p := model.Progress{Pass: pass}
	for i := 0; i <= 20; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		p.Percent = float64(i * 5)
		p.Rescued = fmt.Sprintf("%d%%", i*5)
		p.Rate = "dry-run"
		p.Remaining = "-"
		out <- Update{JobID: jobID, Progress: p, State: model.JobRunning}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func runRealPass(ctx context.Context, job model.Job, pass int, out chan<- Update) error {
	args := buildArgs(pass, job.Src, job.Dst, job.MapFile)
	cmd := exec.CommandContext(ctx, "ddrescue", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	p := model.Progress{Pass: pass}
	errCh := make(chan error, 2)
	var stderrBuf bytes.Buffer
	go func() { errCh <- readProgress(stdout, job.ID, p, out, nil) }()
	go func() { errCh <- readProgress(stderr, job.ID, p, out, &stderrBuf) }()

	waitErr := cmd.Wait()
	_ = <-errCh
	_ = <-errCh
	if waitErr != nil {
		if msg := strings.TrimSpace(stderrBuf.String()); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return waitErr
	}
	return nil
}

// scanLines splits on \n, \r\n, or bare \r (used by ddrescue for in-place updates).
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, b := range data {
		if b == '\r' || b == '\n' {
			j := i + 1
			if j < len(data) && b == '\r' && data[j] == '\n' {
				j++
			}
			return j, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func readProgress(r io.Reader, jobID string, base model.Progress, out chan<- Update, capture *bytes.Buffer) error {
	s := bufio.NewScanner(r)
	s.Split(scanLines)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	p := base
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if capture != nil {
			capture.WriteString(line + "\n")
		}
		p = model.ParseProgressLine(line, p)
		out <- Update{JobID: jobID, Progress: p, State: model.JobRunning}
	}
	return s.Err()
}

func buildArgs(pass int, src, dst, mapFile string) []string {
	switch pass {
	case 1:
		return []string{"--force", "-n", src, dst, mapFile}
	case 2:
		return []string{"--force", "-r3", src, dst, mapFile}
	default:
		return []string{"--force", "-R", src, dst, mapFile}
	}
}
