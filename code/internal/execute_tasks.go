// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Task maintains state and output from various build tasks.
type Task struct {
	Name        string   `json:"name"`
	Language    string   `json:"language"`
	Path        string   `json:"path"`
	Logs        string   `json:"logs"`
	Error       string   `json:"error"`
	Arguments   []string `json:"arguments"`
	err         error
	CmdCreator  func() *exec.Cmd `json:"-"`
	Run         func() error     `json:"-"`
	Cmd         *exec.Cmd        `json:"-"`
	logsBuilder strings.Builder
	Attempts    int           `json:"attempts"`
	Retries     int           `json:"-"`
	Step        BuildStep     `json:"step"`
	Duration    time.Duration `json:"duration"`
}

func (t *Task) String() string {
	const (
		maxPathLength = 60
	)
	path := t.Path
	path = trimFromLeftToLength(path, maxPathLength)

	path = fmt.Sprintf("%-*s", maxPathLength, path)

	return fmt.Sprintf("%-15s%-8s %s", t.Step, t.Language, path)
}

func trimFromLeftToLength(value string, maxLength int) string {
	if len(value) > maxLength {
		value = "..." + value[len(value)-(maxLength-3):]
	}
	return value
}

func (t *Task) run() error {
	ctx, cancel := context.WithTimeout(context.Background(), maxTaskDuration)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.Cmd.Args[0], t.Cmd.Args[1:]...)

	cmd.Dir = t.Path
	cmd.Env = t.Cmd.Env

	t.Cmd = cmd

	var logs []byte
	logs, t.err = t.Cmd.CombinedOutput()
	if t.err != nil {
		t.logsBuilder.WriteString(string(logs))
	}
	return t.err
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
	t.Run = t.run

	return t
}

const maxTaskDuration = 10 * time.Minute
