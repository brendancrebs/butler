// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"selinc.com/butler/code/butler/builtin"
)

// Language specifies options for an individual language defined in the butler config.
type Language struct {
	Name           string              `yaml:"name,omitempty"`
	TaskExec       *TaskCommands       `yaml:"taskCommands,omitempty"`
	DepOptions     *DependencyOptions  `yaml:"dependencyOptions,omitempty"`
	DepCommands    *DependencyCommands `yaml:"dependencyCommands,omitempty"`
	Workspaces     []*Workspace        `yaml:"workspaces,omitempty"`
	StdLibDeps     []string            `yaml:"stdLibDeps,omitempty"`
	ExternalDeps   []string            `yaml:"externalDeps,omitempty"`
	FilePatterns   []string            `yaml:"filePatterns,omitempty"`
	VersionChanged bool                `yaml:"versionChanged,omitempty"`
}

// TaskCommands specifies language specific command options. These will be used to create all of the
// tasks for a language.
type TaskCommands struct {
	SetUp   []string `yaml:"setUp,omitempty"`
	CleanUp []string `yaml:"cleanUp,omitempty"`
	Lint    string   `yaml:"lint,omitempty"`
	Test    string   `yaml:"test,omitempty"`
	Build   string   `yaml:"build,omitempty"`
	Publish string   `yaml:"publish,omitempty"`
}

// DependencyOptions specifies options related to dependency analysis which aren't commands.
type DependencyOptions struct {
	DependencyAnalysis bool `yaml:"dependencyAnalysis,omitempty"`
	ExcludeStdLibs     bool `yaml:"excludeStdLibs,omitempty"`
	ExternalDeps       bool `yaml:"externalDependencies,omitempty"`
}

// DependencyCommands specifies dependency gathering commands for an individual language.
type DependencyCommands struct {
	StandardLibrary string `yaml:"standardLibrary,omitempty"`
	Workspace       string `yaml:"workspace,omitempty"`
	External        string `yaml:"external,omitempty"`
	Version         string `yaml:"version,omitempty"`
}

// Creates tasks for all languages for each build step.
func populateTaskQueue(bc *ButlerConfig, taskQueue *Queue, cmd *cobra.Command) {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "Enumerating repo. Creating build, lint, and test tasks...\n")

	for i := 1; i < len(buildSteps); i++ {
		for _, lang := range bc.Languages {
			buildCommands := getBuildCommands(lang)
			command := buildCommands[buildSteps[i]]
			if command == "" {
				continue
			}
			lang.createTasks(taskQueue, buildSteps[i], command, bc.Task.RunAll)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Done. %s\n\n", time.Since(now).String())
}

// Executes commands that must be run before the creation of tasks
func (lang *Language) runCommandList(cmd *cobra.Command, commands []string) (err error) {
	for _, command := range commands {
		fmt.Fprintf(cmd.OutOrStdout(), "\nexecuting: %s...  ", command)

		commandParts := splitCommand(command)
		if len(commandParts) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "empty command, skipping")
			continue
		}
		command := exec.Command(commandParts[0], commandParts[1:]...)

		if output, err := execOutputStub(command); err != nil {
			return fmt.Errorf("error executing '%s'\nerror: %v\noutput: %v", command, err, output)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "success")
		}
	}
	return
}

// tasks a command as a single string and splits it into multiple parts.
func splitCommand(cmd string) []string {
	commandParts := []string{}
	splitCmd := strings.Fields(cmd)
	commandParts = append(commandParts, splitCmd...)
	return commandParts
}

// executes a command supplied for a user and awaits the response. Any user supplied command is
// expected to pipe an array of strings to stdout as the output of the command.
func ExecuteUserMethods(cmd, name string) (response []string, err error) {
	commandParts := splitCommand(cmd)
	if len(commandParts) == 0 {
		err = fmt.Errorf("dependency commands not supplied for the language %s", name)
		return
	}
	execCmd := exec.Command(commandParts[0], commandParts[1:]...)
	stdout, _ := execCmd.StdoutPipe()

	if err = execStartStub(execCmd); err != nil {
		err = fmt.Errorf("error starting execution of '%s': %v", cmd, err)
		return
	}

	buffer := make([]byte, 1024)
	n, err := readStub(stdout, buffer)
	if err != nil {
		err = fmt.Errorf("error executing '%s': %v", cmd, err)
		return
	}
	responseData := buffer[:n]

	if err = execWaitStub(execCmd); err != nil {
		err = fmt.Errorf("error executing '%s': %v", cmd, err)
		return
	}

	err = json.Unmarshal(responseData, &response)
	return
}

// Gathers all the standard library dependencies and external third party dependencies for a language
// in the repository.
func (lang *Language) getDependencies(bc *ButlerConfig) (err error) {
	if lang.DepOptions.ExcludeStdLibs {
		lang.StdLibDeps, err = builtin.GetStdLibs(lang.Name)
	} else if lang.DepCommands.StandardLibrary != "" {
		lang.StdLibDeps, err = ExecuteUserMethods(lang.DepCommands.StandardLibrary, lang.Name)
	}
	if err != nil {
		return
	}

	if lang.DepOptions.ExternalDeps {
		lang.ExternalDeps, err = builtin.GetExternalDependencies(lang.Name)
	} else {
		lang.ExternalDeps, err = ExecuteUserMethods(lang.DepCommands.External, lang.Name)
	}

	return
}

// Creates a task object for each of a language's workspaces. Each new task will be pushed to the
// task queue.
func (lang *Language) createTasks(taskQueue *Queue, step BuildStep, command string, runAll bool) {
	for _, ws := range lang.Workspaces {
		if ws.IsDirty || runAll {
			command = strings.ReplaceAll(command, "%w", ws.Location)
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

// determines if the options set for a language in the config are valid.
func (lang *Language) validateLanguage() error {
	if lang.Name == "" {
		return errors.New("a language supplied in the config without a name. Please supply a language identifier for each language in the config")
	}
	if len(lang.FilePatterns) < 1 {
		return fmt.Errorf("no file patterns supplied for '%s'. Please see the 'FilePatterns' options in the config spec for more information", lang.Name)
	}

	lang.validateDependencySettings()
	return nil
}

func (lang *Language) validateDependencySettings() {
	if lang.DepCommands == nil {
		lang.DepCommands = &DependencyCommands{
			StandardLibrary: "",
			Workspace:       "",
			External:        "",
		}
	}

	if lang.DepOptions == nil {
		lang.DepOptions = &DependencyOptions{
			DependencyAnalysis: false,
			ExcludeStdLibs:     false,
			ExternalDeps:       false,
		}
	}

	if !lang.DepOptions.DependencyAnalysis {
		lang.DepOptions.ExcludeStdLibs = false
		lang.DepOptions.ExternalDeps = false
		lang.DepCommands.StandardLibrary = ""
		lang.DepCommands.Workspace = ""
		lang.DepCommands.External = ""
	}
}
