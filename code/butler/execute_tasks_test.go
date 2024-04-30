// Copyright (c) 2023 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func testCmdCreator() *exec.Cmd {
	return nil
}

func Test_assignErrorIfError(t *testing.T) {
	type template struct {
		desc     string
		input    error
		next     error
		expected string
	}
	left := errors.New("left")
	right := errors.New("right")
	tests := []*template{
		{"nil with nil", nil, nil, ""},
		{"left with nil", left, nil, "left"},
		{"nil with right", nil, right, "right"},
		{"left with right", left, right, "left; right"},
	}
	for _, test := range tests {
		Convey(test.desc, t, func() {
			err := assignErrorIfError(test.input, test.next)
			if err == nil {
				So(test.expected, ShouldEqual, "")
			} else {
				So(err.Error(), ShouldEqual, test.expected)
			}
		})
	}
}

func Test_trimFromLeft(t *testing.T) {
	type template struct {
		desc     string
		input    string
		expected string
	}
	tests := []*template{
		{"short enough to not be shortened", "short", "short"},
		{"will be shortened", "short short", "...short"},
	}
	for _, test := range tests {
		Convey(test.desc, t, func() {
			result := trimFromLeftToLength(test.input, 8)
			So(result, ShouldEqual, test.expected)
		})
	}
}

func Test_taskReset(t *testing.T) {
	Convey("able to reset a task", t, func() {
		task := createTask("test", "go", "/", 0, 0, testCmdCreator)
		task.Reset()
		So(task.logsBuilder.String(), ShouldEqual, "\n----- TASK RESET -----\n")
	})
}

func failingCommandCreator() *exec.Cmd {
	return &exec.Cmd{
		Path: "zap123",
		Args: []string{"zap123", "build", "-t"},
	}
}

func Test_doWork(t *testing.T) {
	originalStartingSleepDuration := startingSleepDuration
	startingSleepDuration = 10 * time.Millisecond

	defer func() {
		startingSleepDuration = originalStartingSleepDuration
	}()

	Convey("Running a task that restarts a couple times", t, func() {
		task := createTask("test", "go", "/", 0, 2, failingCommandCreator)

		errorChannel := make(chan error, 10)
		const sleepDuration = 10 * time.Millisecond
		var buf bytes.Buffer
		doWork(task, errorChannel, 1, 1, &buf)

		// wait for all threads to finish.
		var err error
		for err == nil {
			select {
			case nextErr := <-errorChannel:
				err = assignErrorIfError(err, nextErr)
			default:
				time.Sleep(sleepDuration)
			}
		}
		So(err.Error(), ShouldEqual, `exec: "zap123": executable file not found in $PATH`)
	})
}
