// Copyright (c) 2023 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
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

const maxTaskDuration = 10 * time.Minute

// run runs the *Task.Cmd and appends output ot *Task.Logs.  Returns any errors.
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

// Reset allows for the task to be Run again.  A task cannot be Run more that once without a reset.
// This appends the reset notice to the logs so that logs are not lost between resets.
func (t *Task) Reset() {
	t.Cmd = t.CmdCreator()
	t.logsBuilder.WriteString("\n----- TASK RESET -----\n")
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

// createTaskName returns `<name> -- <step>`.
// func createTaskName(name, step string) string {
// 	return fmt.Sprintf("%v -- %v", name, step)
// }

// createTask is a constructor for *Task.
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

var numCPU = runtime.NumCPU()

// RunTasksInParallel allows for running a list of tasks in parallel according to the number of
// CPUs visible in the environment.
func RunTasksInParallel(taskQueue *Queue, out io.Writer) (err error) {
	parallelism := int32(numCPU)
	errorChannel := make(chan error, int(parallelism)+1)
	availableThreads := parallelism
	taskNumber := 0
	totalTasks := taskQueue.Size()
	const sleepDuration = 10 * time.Millisecond

	for taskNumber < totalTasks && err == nil {
		select {
		case err = <-errorChannel:
			availableThreads++
		default:
			if availableThreads == 0 {
				time.Sleep(sleepDuration)
				continue
			}
			task := taskQueue.Dequeue()
			taskNumber++
			availableThreads--
			doWork(task, errorChannel, taskNumber, totalTasks, out)
		}
	}
	// wait for all threads to finish.
	for availableThreads < parallelism {
		select {
		case nextErr := <-errorChannel:
			err = assignErrorIfError(err, nextErr)
			availableThreads++
		default:
			time.Sleep(sleepDuration)
		}
	}
	return err
}

func assignErrorIfError(err, nextErr error) error {
	if nextErr == nil {
		return err
	} else if err == nil {
		return nextErr
	}
	return fmt.Errorf("%w; %v", err, nextErr)
}

var startingSleepDuration = time.Second

// doWork is a helper for RunTasksInParallel.  It is guaranteed to add 1 to *availableThreads when
// complete and to return an error on the errorChannel if any.
func doWork(task *Task, errorChannel chan error, taskNumber, totalTasks int, out io.Writer) {
	go func() {
		var err error
		defer func() {
			errorChannel <- err
		}()
		now := time.Now()

		sleepDuration := startingSleepDuration
		err = task.Run()
		task.Attempts++
		for err != nil && task.Retries > 0 {
			task.Retries--
			task.Reset()
			time.Sleep(sleepDuration)
			err = task.Run()
			task.Attempts++
		}
		task.Duration = time.Since(now)
		logPrologue := fmt.Sprintf("%s\n", task.Cmd)
		count := fmt.Sprintf("%s of %d", fmt.Sprintf("%4d", taskNumber), totalTasks)
		attempts := ""
		if task.Attempts > 1 {
			attempts = fmt.Sprintf("(%d)", task.Attempts)
		}
		duration := fmt.Sprintf("%10s ",
			fmt.Sprintf("%s %.1fs", attempts, float64(task.Duration.Nanoseconds())/float64(time.Second)))

		task.Logs = task.logsBuilder.String()
		if err != nil {
			task.Error = err.Error()
			fmt.Fprintf(out, "- %s %s %s %s\n", count, duration, fmt.Sprintf("%75s", task), "FAIL")
			fmt.Fprintf(out, "%s %s\n", logPrologue, task.Logs)
			fmt.Fprintf(out, "Task location: %s\n\n", task.Path)
			return
		}

		fmt.Fprintf(out, "- %s %s %s %s\n", count, duration, fmt.Sprintf("%75s", task), "ok")
	}()
}
