// Copyright (c) 2023 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

// For outputting statistics of the builds.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Results for a blt test run.
type Results struct {
	Error   error  `json:"error"`
	Branch  string `json:"branch"`
	BuildID string `json:"buildID"`
	Commit  string `json:"commit"`

	// Clean if no errors and no Attempts > 1
	// Dirty if no errors and any Attempts > 1
	// Fail of any tasks with errors.
	Status            string        `json:"status"`
	StartTime         time.Time     `json:"startTime"`
	System            *SystemInfo   `json:"system"`
	Tasks             []*Task       `json:"tasks"`
	NestedConfigPaths []string      `json:"nestedConfigPaths"`
	Duration          time.Duration `json:"duration"`
}

// SystemInfo contains info collected and returned as part of butler test results output.
type SystemInfo struct {
	CurrentTime       time.Time `json:"currentTime"`
	Name              string    `json:"name"`
	OperatingSystem   string    `json:"operatingSystem"`
	CPUs              string    `json:"CPUs"`
	Memory            string    `json:"memory"`
	Containers        string    `json:"containers"`
	ContainersRunning string    `json:"containersRunning"`
	Images            string    `json:"images"`
	DiskUsed          string    `json:"diskUsed"`
	DiskSize          string    `json:"diskSize"`
	WorkspaceRoot     string    `json:"workspaceRoot"`
}

var clientDoStub = (*http.Client).Do

// PublishResults serializes results to pretty indented JSON and writes them to the specified filename.
// nolint ignore max arguments per function
func PublishResults(
	cmd *cobra.Command,
	filename string,
	subscribers []string,
	si *SystemInfo,
	tasks []*Task,
	err error,
) error {
	branch, branchErr := getCurrentBranch()
	branch = assignErrorStringIfError(branch, branchErr)

	results := &Results{
		Branch:    branch,
		BuildID:   GetEnvOrDefault(envBuildID, "No BUILD_ID environment variable"),
		Commit:    GetEnvOrDefault(envBitbucketCommit, "No GIT_COMMIT environment variable"),
		Status:    getStatusFromTasks(tasks).String(),
		StartTime: si.CurrentTime,
		Duration:  time.Since(si.CurrentTime),
		System:    si,
		Tasks:     tasks,
		Error:     err,
	}

	// not possible to error here as we control everything.
	prettyBytes, _ := json.MarshalIndent(results, "", "\t")

	fmt.Fprintf(cmd.OutOrStdout(), "Notifying subscribers...\n")

	// send results to subscribers
	for _, subscriber := range subscribers {
		notifySubscriber(cmd, prettyBytes, subscriber)
	}

	return os.WriteFile(filename, prettyBytes, 0o600)
}

func notifySubscriber(cmd *cobra.Command, results []byte, subscriberURI string) {
	request, _ := http.NewRequest("POST", subscriberURI, bytes.NewBuffer(results))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := clientDoStub(client, request)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Received error from subscriber: %s\n", err)
	} else {
		response.Body.Close()
	}
}

func assignErrorStringIfError(value string, err error) string {
	if err == nil {
		return value
	}
	return err.Error()
}

func getStatusFromTasks(tasks []*Task) BuildStatus {
	status := BuildStatusClean
	for _, task := range tasks {
		if task.Attempts > 1 {
			status = BuildStatusDirty
		}
		if task.err != nil {
			status = BuildStatusFail
			break
		}
	}
	return status
}

// ErrUnexpectedDFOutput is exported error.
var ErrUnexpectedDFOutput = errors.New("unexpected df output")

// GetSystemInfo queries docker, df and free commands to collect various system stats.
func GetSystemInfo() (si SystemInfo, err error) {
	si = SystemInfo{}
	si.CurrentTime = time.Now()

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return
	}

	dockerCmd := &exec.Cmd{
		Path: dockerPath,
		Args: []string{dockerPath, "system", "info", "--format", `"{{json . }}"`},
	}
	dsi, err := dockerCmd.Output()
	if err != nil {
		return
	}

	dfPath, err := exec.LookPath("df")
	if err != nil {
		return
	}

	si.Containers = getValueFromBytes(dsi, `"Containers":`)
	si.ContainersRunning = getValueFromBytes(dsi, `"ContainersRunning":`)
	si.Images = getValueFromBytes(dsi, `"Images":`)
	si.OperatingSystem = getValueFromBytes(dsi, `"OperatingSystem":`)
	si.Memory = getValueFromBytes(dsi, `"MemTotal":`)
	si.CPUs = getValueFromBytes(dsi, `"NCPU":`)
	si.Name = getValueFromBytes(dsi, `,"Name":`)
	si.DiskSize = getDFOutputByField(dfPath, "/", "size")
	si.DiskUsed = getDFOutputByField(dfPath, "/", "used")

	return
}

// getDFOutputByField gets field output from the df command for root, e.g. "/".
func getDFOutputByField(dfPath, path, field string) string {
	dfCmd := &exec.Cmd{Path: dfPath, Args: []string{dfPath, path, fmt.Sprintf("--output=%s", field)}}
	output, err := dfCmd.Output()
	if err != nil {
		return err.Error()
	}
	parts := bytes.Split(bytes.TrimSpace(output), []byte("\n")) // split lines
	if len(parts) != 2 {
		return ErrUnexpectedDFOutput.Error()
	}
	return strings.TrimSpace(string(parts[1]))
}

// getValueFromBytes takes an input key to split on, if split is successful, returns the value to
// the right of split, but before the next comma.
func getValueFromBytes(b []byte, key string) (value string) {
	parts := bytes.Split(b, []byte(key))
	if len(parts) > 1 {
		value = strings.Trim(string(bytes.Split(parts[1], []byte(","))[0]), `"`)
	}
	return
}
