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
	Name                      string              `yaml:"Name,omitempty"`
	FileExtension             string              `yaml:"FileExtension,omitempty"`
	VersionPath               string              `yaml:"VersionPath,omitempty"`
	WorkspaceFile             string              `yaml:"WorkspaceFile,omitempty"`
	TaskExec                  *TaskCommands       `yaml:"TaskCommands,omitempty"`
	DepCommands               *DependencyCommands `yaml:"DependencyCommands,omitempty"`
	Workspaces                []*Workspace        `yaml:"Workspaces,omitempty"`
	StdLibDeps                []string            `yaml:"StdLibDeps,omitempty"`
	ExternalDeps              []string            `yaml:"ExternalDeps,omitempty"`
	BuiltinStdLibsMethod      bool                `yaml:"BuiltinStdLibsMethod,omitempty"`
	BuiltinWorkspaceDepMethod bool                `yaml:"BuiltinWorkspaceDepMethod,omitempty"`
	BuiltinExternalDepMethod  bool                `yaml:"BuiltinExternalDepMethod,omitempty"`
	VersionChanged            bool                `yaml:"VersionChanged,omitempty"`
}

type TaskCommands struct {
	SetUpCommands  []string `yaml:"SetUpCommands,omitempty"`
	LintPath       string   `yaml:"LintPath,omitempty"`
	LintCommand    string   `yaml:"LintCommand,omitempty"`
	TestPath       string   `yaml:"TestPath,omitempty"`
	TestCommand    string   `yaml:"TestCommand,omitempty"`
	BuildPath      string   `yaml:"BuildMethodPath,omitempty"`
	BuildCommand   string   `yaml:"BuildCommand,omitempty"`
	PublishPath    string   `yaml:"PublishPath,omitempty"`
	PublishCommand string   `yaml:"PublishCommand,omitempty"`
}

type DependencyCommands struct {
	StdLibsPath         string `yaml:"stdLibsPath,omitempty"`
	StdLibsCommand      string `yaml:"stdLibsCommand,omitempty"`
	WorkspaceDepPath    string `yaml:"InternalDepPath,omitempty"`
	WorkspaceDepCommand string `yaml:"InternalDepCommand,omitempty"`
	ExternalDepPath     string `yaml:"ExternalDepPath,omitempty"`
	ExternalDepCommand  string `yaml:"ExternalDepCommand,omitempty"`
}

func PopulateTaskQueue(bc *ButlerConfig, taskQueue *Queue, cmd *cobra.Command) {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "Enumerating repo. Creating build, lint, and test tasks...\n")

	// The calls for createTasks are in separate loops so lint tasks for all languages will be first
	// in queue and so on for the other task types.
	for _, lang := range bc.Languages {
		createTasks(lang, taskQueue, BuildStepLint, lang.TaskExec.LintCommand, lang.TaskExec.LintPath)
	}
	for _, lang := range bc.Languages {
		createTasks(lang, taskQueue, BuildStepTest, lang.TaskExec.TestCommand, lang.TaskExec.TestPath)
	}
	for _, lang := range bc.Languages {
		createTasks(lang, taskQueue, BuildStepBuild, lang.TaskExec.BuildCommand, lang.TaskExec.BuildPath)
	}
	for _, lang := range bc.Languages {
		createTasks(lang, taskQueue, BuildStepPublish, lang.TaskExec.PublishCommand, lang.TaskExec.PublishPath)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Done. %s\n\n", time.Since(now).String())
}

// Executes commands that must be run before the creation of tasks
func preliminaryCommands(langs []*Language) (err error) {
	for _, lang := range langs {
		for _, cmd := range lang.TaskExec.SetUpCommands {
			fmt.Printf("\nexecuting: %s...  ", cmd)

			commandParts := splitCommand(cmd)
			if len(commandParts) == 0 {
				fmt.Println("empty command, skipping")
				continue
			}
			cmd := exec.Command(commandParts[0], commandParts[1:]...)

			if output, err := execOutputStub(cmd); err != nil {
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

func executeUserMethods(cmd, path, name string) (response []string, err error) {
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

	if err = execStartStub(execCmd); err != nil {
		err = fmt.Errorf("error starting execution of '%s': %v\n", cmd, err)
		return
	}

	buffer := make([]byte, 1024)
	n, err := stdout.Read(buffer)
	if err != nil {
		return
	}
	responseData := buffer[:n]

	if err = execWaitStub(execCmd); err != nil {
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
		lang.StdLibDeps, err = executeUserMethods(lang.DepCommands.ExternalDepCommand, lang.DepCommands.ExternalDepPath, lang.Name)
	}
	if err != nil {
		return
	}

	return
}

func createTasks(lang *Language, taskQueue *Queue, step BuildStep, command, commandPath string) {
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
