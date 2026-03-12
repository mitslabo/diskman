package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type JSONL struct {
	mu   sync.Mutex
	file *os.File
}

func New(path string) (*JSONL, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &JSONL{file: f}, nil
}

func (l *JSONL) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *JSONL) Log(v any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = l.file.Write(append(b, '\n'))
	return err
}
