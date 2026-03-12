package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type JobState string

const (
	JobPending   JobState = "pending"
	JobRunning   JobState = "running"
	JobDone      JobState = "done"
	JobError     JobState = "error"
	JobCancelled JobState = "cancelled"
)

type Progress struct {
	Pass      int
	Percent   float64
	Rescued   string
	Rate      string
	Remaining string
	BadAreas  int
	ReadErrs  int
}

type Job struct {
	ID        string
	Op        string
	Name      string
	Src       string
	Dst       string
	MapFile   string
	State     JobState
	Progress  Progress
	ErrMsg    string
	CreatedAt time.Time
}

func NewJobID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(b)
}
