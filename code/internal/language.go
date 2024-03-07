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

	"selinc.com/butler/code/internal/builtin"
)

type Language struct {
	Name                      string              `yaml:"name,omitempty"`
	FileExtension             string              `yaml:"fileExtension,omitempty"`
	VersionPath               string              `yaml:"versionPath,omitempty"`
	WorkspaceFile             string              `yaml:"workspaceFile,omitempty"`
	TaskExec                  *TaskCommands       `yaml:"taskCommands,omitempty"`
	DepCommands               *DependencyCommands `yaml:"dependencyCommands,omitempty"`
	Workspaces                []*Workspace        `yaml:"workspaces,omitempty"`
	StdLibDeps                []string            `yaml:"stdLibDeps,omitempty"`
	ExternalDeps              []string            `yaml:"externalDeps,omitempty"`
	BuiltinStdLibsMethod      bool                `yaml:"builtinStdLibsMethod,omitempty"`
	BuiltinWorkspaceDepMethod bool                `yaml:"builtinWorkspaceDepMethod,omitempty"`
	BuiltinExternalDepMethod  bool                `yaml:"builtinExternalDepMethod,omitempty"`
	VersionChanged            bool                `yaml:"versionChanged,omitempty"`
}

type TaskCommands struct {
	SetUpCommands  []string `yaml:"setUpCommands,omitempty"`
	LintPath       string   `yaml:"lintPath,omitempty"`
	LintCommand    string   `yaml:"lintCommand,omitempty"`
	TestPath       string   `yaml:"testPath,omitempty"`
	TestCommand    string   `yaml:"testCommand,omitempty"`
	BuildPath      string   `yaml:"buildPath,omitempty"`
	BuildCommand   string   `yaml:"buildCommand,omitempty"`
	PublishPath    string   `yaml:"publishPath,omitempty"`
	PublishCommand string   `yaml:"publishCommand,omitempty"`
}

type DependencyCommands struct {
	StdLibsPath         string `yaml:"stdLibsPath,omitempty"`
	StdLibsCommand      string `yaml:"stdLibsCommand,omitempty"`
	WorkspaceDepPath    string `yaml:"internalDepPath,omitempty"`
	WorkspaceDepCommand string `yaml:"internalDepCommand,omitempty"`
	ExternalDepPath     string `yaml:"externalDepPath,omitempty"`
	ExternalDepCommand  string `yaml:"externalDepCommand,omitempty"`
}

func PopulateTaskQueue(bc *ButlerConfig, taskQueue *Queue, cmd *cobra.Command) {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "Enumerating repo. Creating build, lint, and test tasks...\n")

	// The calls for createTasks are in separate loops so lint tasks for all languages will be first
	// in queue and so on for the other task types.
	for _, lang := range bc.Languages {
		CreateTasks(lang, taskQueue, BuildStepLint, lang.TaskExec.LintCommand, lang.TaskExec.LintPath)
	}
	for _, lang := range bc.Languages {
		CreateTasks(lang, taskQueue, BuildStepTest, lang.TaskExec.TestCommand, lang.TaskExec.TestPath)
	}
	for _, lang := range bc.Languages {
		CreateTasks(lang, taskQueue, BuildStepBuild, lang.TaskExec.BuildCommand, lang.TaskExec.BuildPath)
	}
	for _, lang := range bc.Languages {
		CreateTasks(lang, taskQueue, BuildStepPublish, lang.TaskExec.PublishCommand, lang.TaskExec.PublishPath)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Done. %s\n\n", time.Since(now).String())
}

// Executes commands that must be run before the creation of tasks
func PreliminaryCommands(langs []*Language) (err error) {
	for _, lang := range langs {
		for _, cmd := range lang.TaskExec.SetUpCommands {
			fmt.Printf("\nexecuting: %s...  ", cmd)

			commandParts := splitCommand(cmd)
			if len(commandParts) == 0 {
				fmt.Println("empty command, skipping")
				continue
			}
			cmd := exec.Command(commandParts[0], commandParts[1:]...)

			if output, err := ExecOutputStub(cmd); err != nil {
				err = fmt.Errorf("error executing '%s':\nerror: %v\noutput: %v", cmd, err, output)
				return err
			} else {
				fmt.Printf("success\n")
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

func ExecuteUserMethods(cmd, path, name string) (response []string, err error) {
	commandParts := splitCommand(cmd)
	if len(commandParts) == 0 {
		err = fmt.Errorf("dependency commands not supplied for the language %s.\n", name)
		return
	}
	execCmd := exec.Command(commandParts[0], commandParts[1:]...)
	if path != "" {
		execCmd.Dir = path
	}
	stdout, _ := execCmd.StdoutPipe()

	if err = ExecStartStub(execCmd); err != nil {
		err = fmt.Errorf("error starting execution of '%s': %v\n", cmd, err)
		return
	}

	buffer := make([]byte, 1024)
	n, err := stdout.Read(buffer)
	if err != nil {
		return
	}
	responseData := buffer[:n]

	if err = ExecWaitStub(execCmd); err != nil {
		err = fmt.Errorf("error executing '%s': %v\n", cmd, err)
		return
	}

	if err = json.Unmarshal(responseData, &response); err != nil {
		return
	}
	return
}

func (lang *Language) getExternalDeps(bc *ButlerConfig) (err error) {
	lang.Name, err = builtin.GetLanguageId(lang.Name)
	if err != nil {
		return
	}

	if lang.BuiltinStdLibsMethod {
		lang.StdLibDeps, err = builtin.GetStdLibs(lang.Name)
	} else {
		lang.StdLibDeps, err = ExecuteUserMethods(lang.DepCommands.ExternalDepCommand, lang.DepCommands.ExternalDepPath, lang.Name)
	}
	if err != nil {
		return
	}

	return
}

func CreateTasks(lang *Language, taskQueue *Queue, step BuildStep, command, commandPath string) {
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
}

// formats command strings from the butler config
func replaceSubstring(input string, substring string, path string) string {
	return strings.ReplaceAll(input, substring, path)
}
