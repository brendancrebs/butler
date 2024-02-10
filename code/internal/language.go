// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"selinc.com/butler/code/builtin"

	"github.com/spf13/cobra"
)

type Language struct {
	Name                     string       `yaml:"Name,omitempty"`
	FileExtension            string       `yaml:"FileExtension,omitempty"`
	PackageManager           string       `yaml:"PackageManager,omitempty"`
	VersionPath              string       `yaml:"VersionPath,omitempty"`
	WorkspaceFile            string       `yaml:"WorkspaceFile,omitempty"`
	UserCommands             *Commands    `yaml:"Commands,omitempty"`
	Workspaces               []*Workspace `yaml:"Workspaces,omitempty"`
	SetUpCommands            []string     `yaml:"SetUpCommands,omitempty"`
	ChangedDeps              []string     `yaml:"ChangedDeps,omitempty"`
	DefaultExternalDepMethod bool         `yaml:"DefaultExternalDepMethod,omitempty"`
	VersionChanged           bool         `yaml:"VersionChanged,omitempty"`
}

type Commands struct {
	LintPath           string `yaml:"LintPath,omitempty"`
	LintCommand        string `yaml:"LintCommand,omitempty"`
	TestPath           string `yaml:"TestPath,omitempty"`
	TestCommand        string `yaml:"TestCommand,omitempty"`
	BuildPath          string `yaml:"BuildMethodPath,omitempty"`
	BuildCommand       string `yaml:"BuildCommand,omitempty"`
	PublishPath        string `yaml:"PublishPath,omitempty"`
	PublishCommand     string `yaml:"PublishCommand,omitempty"`
	ExternalDepPath    string `yaml:"ExternalDepPath,omitempty"`
	ExternalDepCommand string `yaml:"ExternalDepCommand,omitempty"`
}

func PopulateTaskQueue(bc *ButlerConfig, taskQueue *Queue, cmd *cobra.Command) (err error) {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "Enumerating repo. Creating build, lint, and test tasks...\n")
	if bc.Task.ShouldLint {
		if err = createTasks(bc, taskQueue, BuildStepLint); err != nil {
			return fmt.Errorf("Error creating lint tasks: %v\n", err)
		}
	}
	if bc.Task.ShouldTest {
		if err = createTasks(bc, taskQueue, BuildStepTest); err != nil {
			return fmt.Errorf("Error creating test tasks: %v\n", err)
		}
	}
	if bc.Task.ShouldBuild {
		if err = createTasks(bc, taskQueue, BuildStepBuild); err != nil {
			return fmt.Errorf("Error creating build tasks: %v\n", err)
		}
	}
	if bc.Task.ShouldPublish {
		if err = createTasks(bc, taskQueue, BuildStepPublish); err != nil {
			return fmt.Errorf("Error creating publish tasks: %v\n", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Done. %s\n\n", time.Since(now).String())

	return
}

// Executes commands that must be run before the creation of tasks
func preliminaryCommands(langs []*Language) (err error) {
	for _, lang := range langs {
		for _, cmd := range lang.SetUpCommands {
			fmt.Printf("\nExecuting: %s...  ", cmd)

			commandParts := splitCommand(cmd)
			if len(commandParts) == 0 {
				fmt.Println("Empty command, skipping")
				continue
			}

			execCmd := exec.Command(commandParts[0], commandParts[1:]...)
			// execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err = execCmd.Run(); err != nil {
				err = fmt.Errorf("Error executing '%s':\n %v\n", cmd, err)
			} else {
				fmt.Printf("Success\n")
			}
		}
	}
	return
}

func splitCommand(cmd string) []string {
	commandParts := []string{}
	splitCmd := strings.Fields(cmd)
	commandParts = append(commandParts, splitCmd...)
	return commandParts
}

func executeUserMethods(cmd, path, step, name string, marshaledData []byte) (responseData []byte, err error) {
	commandParts := splitCommand(cmd)
	if len(commandParts) == 0 {
		err = fmt.Errorf("%s command not supplied for the language %s.\n", step, name)
		return
	}
	execCmd := exec.Command(commandParts[0], commandParts[1:]...)
	if path != "" {
		execCmd.Dir = path
	}
	stdin, _ := execCmd.StdinPipe()
	stdout, _ := execCmd.StdoutPipe()

	if err = execCmd.Start(); err != nil {
		err = fmt.Errorf("Error executing '%s':\n %v\n", cmd, err)
		return
	}

	fmt.Fprintln(stdin, string(marshaledData))
	stdin.Close()

	buffer := make([]byte, 1024)
	n, err := stdout.Read(buffer)
	if err != nil {
		err = fmt.Errorf("Error reading response for user %s method: %v", step, err)
		return
	}
	responseData = buffer[:n]

	if err = execCmd.Wait(); err != nil {
		err = fmt.Errorf("Error executing '%s':\n %v\n", cmd, err)
		return
	}
	return
}

func CreateWorkspaces(bc *ButlerConfig, paths []string) (err error) {
	for _, lang := range bc.Languages {
		lang.Workspaces, err = getWorkspaces(bc, lang, paths)
		if err != nil {
			return fmt.Errorf("Error getting workspaces for %s: %v\n", lang.Name, err)
		}

		if !lang.DefaultExternalDepMethod {
			lang.ChangedDeps, err = getExternalDeps(lang.UserCommands, lang.Name)
		} else {
			lang.ChangedDeps, err = builtInGetThirdPartyDeps(lang.Name, bc.Paths.WorkspaceRoot)
		}
		if err != nil {
			return fmt.Errorf("Error getting external dependencies for %s: %v\n", lang.Name, err)
		}
	}
	return
}

func getExternalDeps(Commands *Commands, name string) (changedDeps []string, err error) {
	buffer, err := executeUserMethods(Commands.ExternalDepCommand, Commands.ExternalDepPath, "external dependency", name, []byte{})
	if err != nil {
		return
	}

	if err = json.Unmarshal(buffer, &changedDeps); err != nil {
		err = fmt.Errorf("Error unmarshaling changed external dependencies: %v\n", err)
		return
	}

	return
}

func builtInGetThirdPartyDeps(name string, workspaceRoot string) (changedDeps []string, err error) {
	changedDeps, err = builtin.ExternalDependencyParsing(name, workspaceRoot)
	if err != nil {
		return
	}
	return
}

func createTasks(bc *ButlerConfig, taskQueue *Queue, step BuildStep) (err error) {
	for _, lang := range bc.Languages {
		command, commandPath, err := getCommandByStep(step, lang)
		if err != nil {
			break
		}
		for _, ws := range lang.Workspaces {
			if bc.Task.ShouldRunAll || ws.IsDirty {
				command = replaceSubstring(command, "%w", ws.Location)
				if strings.Contains(command, "%p") {
					commandPath, _ = executeCommandIfExists(commandPath)
					command = replaceSubstring(command, "%p", commandPath)
				}
				createCmd := func() *exec.Cmd {
					return &exec.Cmd{
						Path: ws.Location,
						Args: strings.Fields(command),
					}
				}
				taskQueue.Enqueue(createTask(ws.Location, lang.Name, ws.Location, 0, step, createCmd))
			}
		}
	}
	return
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

// formats command strings from the butler config
func replaceSubstring(input string, substring string, path string) string {
	return strings.ReplaceAll(input, substring, path)
}

func executeCommandIfExists(input string) (string, error) {
	commandSuffix := "--command"

	// Check if the input string ends with "--command"
	if strings.HasSuffix(input, commandSuffix) {
		commandStr := strings.TrimSpace(strings.TrimSuffix(input, commandSuffix))

		cmd := exec.Command("sh", "-c", commandStr)
		outputBytes, err := cmd.Output()
		if err != nil {
			return "", err
		}

		return string(outputBytes), nil
	}

	return input, nil
}

func getCommandByStep(step BuildStep, lang *Language) (string, string, error) {
	switch step {
	case BuildStepLint:
		return lang.UserCommands.LintCommand, lang.UserCommands.LintPath, nil
	case BuildStepTest:
		return lang.UserCommands.TestCommand, lang.UserCommands.TestPath, nil
	case BuildStepBuild:
		return lang.UserCommands.BuildCommand, lang.UserCommands.BuildPath, nil
	case BuildStepPublish:
		return lang.UserCommands.PublishCommand, lang.UserCommands.PublishPath, nil
	default:
		return "", "", fmt.Errorf("unknown build step\n")
	}
}
