// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"os/exec"
	"strings"
	"time"
)

may just be showing how much I don't know about butler in general, but is there a reason you need to store both Attempts and Retries? Could Retries not be generated if Attempts > 1?
// Task maintains state and output from various build tasks.
type Task struct {
	Name        string           `json:"name"`
	Language    string           `json:"language"`
	Path        string           `json:"path"`
	Logs        string           `json:"logs"`
	Error       string           `json:"error"`
	Arguments   []string         `json:"arguments"`
	CmdCreator  func() *exec.Cmd `json:"-"`
	Run         func() error     `json:"-"`
	Cmd         *exec.Cmd        `json:"-"`
	logsBuilder strings.Builder
	Attempts    int           `json:"attempts"`
	Retries     int           `json:"-"`
	Step        BuildStep     `json:"step"`
	Duration    time.Duration `json:"duration"`
}

func createTask(name, language, path string, retries int, step BuildStep, createCmd func() *exec.Cmd) *Task {
	t := &Task{
		Name:        name,
		Path:        path,
		Language:    language,
		Step:        step,
		CmdCreator:  createCmd,
		Cmd:         createCmd(),
		logsBuilder: strings.Builder{},
		Retries:     retries,
	}
	return t
}
