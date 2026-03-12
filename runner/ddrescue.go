package runner

import (
	"bufio"
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
		job.State = model.JobRunning
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
	readFn := func(r io.Reader) {
		errCh <- readProgress(r, job.ID, p, out)
	}
	go readFn(stdout)
	go readFn(stderr)

	waitErr := cmd.Wait()
	_ = <-errCh
	_ = <-errCh
	if waitErr != nil {
		return waitErr
	}
	return nil
}

func readProgress(r io.Reader, jobID string, base model.Progress, out chan<- Update) error {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	p := base
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		p = model.ParseProgressLine(line, p)
		out <- Update{JobID: jobID, Progress: p, State: model.JobRunning}
	}
	return s.Err()
}

func buildArgs(pass int, src, dst, mapFile string) []string {
	switch pass {
	case 1:
		return []string{"-n", src, dst, mapFile}
	case 2:
		return []string{"-r3", src, dst, mapFile}
	default:
		return []string{"-R", src, dst, mapFile}
	}
}
