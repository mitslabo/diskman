package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"diskman/model"
)

var (
	reDdBytes = regexp.MustCompile(`^(\d+) bytes`)
	reDdRate  = regexp.MustCompile(`,\s+([\d.]+ \S+/s)`)
)

// deviceSize returns the byte size of a block device or file by seeking to end.
func deviceSize(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Seek(0, io.SeekEnd)
}

func runErase(ctx context.Context, job model.Job, out chan<- Update) error {
	totalBytes, _ := deviceSize(job.Dst)

	cmd := exec.CommandContext(ctx, "dd",
		"if=/dev/zero",
		"of="+job.Dst,
		"bs=1M",
		"status=progress",
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	p := model.Progress{Pass: 1}
	var stderrBuf bytes.Buffer
	errCh := make(chan error, 1)
	go func() {
		errCh <- readEraseProgress(stderr, job.ID, p, totalBytes, out, &stderrBuf)
	}()

	waitErr := cmd.Wait()
	_ = <-errCh

	if waitErr != nil {
		stderrText := stderrBuf.String()
		// "No space left on device" = device fully zeroed — treat as success
		if strings.Contains(stderrText, "No space left on device") {
			return nil
		}
		// return the dd error line if present
		for _, line := range strings.Split(stderrText, "\n") {
			if strings.HasPrefix(line, "dd: ") {
				return fmt.Errorf("%s", line)
			}
		}
		return waitErr
	}
	return nil
}

func readEraseProgress(r io.Reader, jobID string, base model.Progress, totalBytes int64, out chan<- Update, stderrBuf *bytes.Buffer) error {
	s := bufio.NewScanner(r)
	s.Split(scanLines)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	p := base
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		stderrBuf.WriteString(line + "\n")
		p = parseDdProgress(line, p, totalBytes)
		out <- Update{JobID: jobID, Progress: p, State: model.JobRunning}
	}
	return s.Err()
}

func parseDdProgress(line string, prev model.Progress, totalBytes int64) model.Progress {
	p := prev
	if m := reDdBytes.FindStringSubmatch(line); len(m) == 2 {
		if written, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			p.Rescued = formatBytes(written)
			if totalBytes > 0 {
				p.Percent = float64(written) / float64(totalBytes) * 100
				if p.Percent > 100 {
					p.Percent = 100
				}
			}
		}
	}
	if m := reDdRate.FindStringSubmatch(line); len(m) == 2 {
		p.Rate = strings.TrimSpace(m[1])
	}
	return p
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func runDryErase(ctx context.Context, jobID string, out chan<- Update) error {
	p := model.Progress{Pass: 1}
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
