// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package internal

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"selinc.com/butler/code/builtin"
)

type Language struct {
	Name                     string       `yaml:"Name,omitempty"`
	FileExtension            string       `yaml:"FileExtension,omitempty"`
	PackageManager           string       `yaml:"PackageManager,omitempty"`
	VersionPath              string       `yaml:"VersionPath,omitempty"`
	WorkspaceFile            string       `yaml:"WorkspaceFile,omitempty"`
	Commands                 *Commands    `yaml:"Commands,omitempty"`
	Workspaces               []*Workspace `yaml:"Workspaces,omitempty"`
	ChangedDeps              []string     `yaml:"ChangedDeps,omitempty"`
	DefaultExternalDepMethod bool         `yaml:"DefaultExternalDepMethod,omitempty"`
	VersionChanged           bool         `yaml:"VersionChanged,omitempty"`
}

type Commands struct {
	SetUpCommands      []string `yaml:"SetUpCommands,omitempty"`
	LintPath           string   `yaml:"LintPath,omitempty"`
	LintCommand        string   `yaml:"LintCommand,omitempty"`
	TestPath           string   `yaml:"TestPath,omitempty"`
	TestCommand        string   `yaml:"TestCommand,omitempty"`
	BuildPath          string   `yaml:"BuildMethodPath,omitempty"`
	BuildCommand       string   `yaml:"BuildCommand,omitempty"`
	PublishPath        string   `yaml:"PublishPath,omitempty"`
	PublishCommand     string   `yaml:"PublishCommand,omitempty"`
	ExternalDepPath    string   `yaml:"ExternalDepPath,omitempty"`
	ExternalDepCommand string   `yaml:"ExternalDepCommand,omitempty"`
}

func PopulateTaskQueue(bc *ButlerConfig, taskQueue *Queue, cmd *cobra.Command) (err error) {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "Enumerating repo. Creating build, lint, and test tasks...\n")
	for _, lang := range bc.Languages {
		if err = createTasks(lang, taskQueue, BuildStepLint, lang.Commands.LintCommand, lang.Commands.LintPath); err != nil {
			return fmt.Errorf("Error creating lint tasks: %v\n", err)
		}
	}
	for _, lang := range bc.Languages {
		if err = createTasks(lang, taskQueue, BuildStepTest, lang.Commands.TestCommand, lang.Commands.TestPath); err != nil {
			return fmt.Errorf("Error creating Test tasks: %v\n", err)
		}
	}
	for _, lang := range bc.Languages {
		if err = createTasks(lang, taskQueue, BuildStepBuild, lang.Commands.BuildCommand, lang.Commands.BuildPath); err != nil {
			return fmt.Errorf("Error creating build tasks: %v\n", err)
		}
	}
	for _, lang := range bc.Languages {
		if err = createTasks(lang, taskQueue, BuildStepPublish, lang.Commands.PublishCommand, lang.Commands.PublishPath); err != nil {
			return fmt.Errorf("Error creating publish tasks: %v\n", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Done. %s\n\n", time.Since(now).String())

	return
}

// Executes commands that must be run before the creation of tasks
func preliminaryCommands(langs []*Language) (err error) {
	for _, lang := range langs {
		for _, cmd := range lang.Commands.SetUpCommands {
			fmt.Printf("\nExecuting: %s...  ", cmd)

			commandParts := splitCommand(cmd)
			if len(commandParts) == 0 {
				fmt.Println("Empty command, skipping")
				continue
			}

			cmd := exec.Command(commandParts[0], commandParts[1:]...)

			if output, err := execOutputStub(cmd); err != nil {
				err = fmt.Errorf("Error executing '%s':\nError: %v\nOutput: %v", cmd, err, output)
				return err
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
			lang.ChangedDeps, err = getExternalDeps(lang.Commands, lang.Name)
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

func createTasks(lang *Language, taskQueue *Queue, step BuildStep, command, commandPath string) (err error) {
	for _, ws := range lang.Workspaces {
		if ws.IsDirty {
			command = replaceSubstring(command, "%w", ws.Location)
			createCmd := func() *exec.Cmd {
				return &exec.Cmd{
					Path: ws.Location,
					Args: strings.Fields(command),
				}
			}
			taskQueue.Enqueue(createTask(ws.Location, lang.Name, ws.Location, 0, step, createCmd))
		}
	}
	return
}

// formats command strings from the butler config
func replaceSubstring(input string, substring string, path string) string {
	return strings.ReplaceAll(input, substring, path)
}
