// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"selinc.com/butler/code/butler/builtin"
)

// Specifies options for an individual language defined in the butler config.
type Language struct {
	Name           string              `yaml:"name,omitempty"`
	TaskOptions    *TaskConfigurations `yaml:"taskOptions,omitempty"`
	TaskCmd        *TaskCommands       `yaml:"taskCommands,omitempty"`
	DepOptions     *DependencyOptions  `yaml:"dependencyOptions,omitempty"`
	DepCommands    *DependencyCommands `yaml:"dependencyCommands,omitempty"`
	Workspaces     []*Workspace        `yaml:"workspaces,omitempty"`
	StdLibDeps     []string            `yaml:"stdLibDeps,omitempty"`
	ExternalDeps   []string            `yaml:"externalDeps,omitempty"`
	WorkspaceFiles []string            `yaml:"workspaceFiles,omitempty"`
}

// Specifies language specific command options. These will be used to create all of the tasks for a
// language.
type TaskCommands struct {
	SetUp   []string `yaml:"setUp,omitempty"`
	CleanUp []string `yaml:"cleanUp,omitempty"`
	Lint    string   `yaml:"lint,omitempty"`
	Test    string   `yaml:"test,omitempty"`
	Build   string   `yaml:"build,omitempty"`
	Publish string   `yaml:"publish,omitempty"`
}

// Specifies options related to dependency analysis which aren't commands.
type DependencyOptions struct {
	DependencyAnalysis bool `yaml:"dependencyAnalysis,omitempty"`
	ExcludeStdLibs     bool `yaml:"excludeStdLibs,omitempty"`
	ExternalDeps       bool `yaml:"externalDependencies,omitempty"`
}

// Specifies dependency gathering commands for an individual language.
type DependencyCommands struct {
	StandardLibrary string `yaml:"standardLibrary,omitempty"`
	Workspace       string `yaml:"workspace,omitempty"`
	External        string `yaml:"external,omitempty"`
}

// Creates tasks for all languages and populates the task queue.
func populateTaskQueue(bc *ButlerConfig, taskQueue *Queue, cmd *cobra.Command) {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "Enumerating repo. Creating build, lint, and test tasks...\n")

	buildEnabled := getEnabledBuildSteps(bc)
	for i := 1; i < len(buildSteps); i++ {
		if !buildEnabled[buildSteps[i]] && !bc.Task.RunAll {
			continue
		}
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

// Runs a list of shell commands one after the other. If any command fails, Butler will fail.
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

// Takes a command as a single string and splits it on spaces.
func splitCommand(cmd string) []string {
	commandParts := []string{}
	splitCmd := strings.Fields(cmd)
	commandParts = append(commandParts, splitCmd...)
	return commandParts
}

// Executes a command supplied for a user and awaits the response. Any user supplied command is
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

	buffer := make([]byte, 8192)
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

// Gathers all the standard library dependencies and external third party dependencies for a
// language in the repository.
func (lang *Language) getDependencies(bc *ButlerConfig) (err error) {
	if lang.DepOptions.ExcludeStdLibs {
		lang.StdLibDeps, _ = builtin.GetStdLibs(lang.Name)
	} else if lang.DepCommands.StandardLibrary != "" {
		lang.StdLibDeps, err = ExecuteUserMethods(lang.DepCommands.StandardLibrary, lang.Name)
	}
	if err != nil {
		return
	}

	if len(lang.StdLibDeps) > 0 {
		versionChanged, err := strconv.ParseBool(lang.StdLibDeps[0])
		if err == nil {
			bc.Task.RunAll = bc.Task.RunAll || versionChanged
		}
	}

	if lang.DepOptions.ExternalDeps {
		lang.ExternalDeps, _ = builtin.GetExternalDependencies(lang.Name)
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

// Determines if the options set for a language in the config are valid.
func (lang *Language) validateLanguage(bc *ButlerConfig) error {
	if lang.Name == "" {
		return errors.New("a language was supplied in the config without a name. Please supply a language identifier for each language in the config")
	}
	if len(lang.WorkspaceFiles) < 1 {
		return fmt.Errorf("no file patterns supplied for '%s'. Please see the 'workspaceFiles' options in the language_options.md spec for more information", lang.Name)
	}

	lang.validateDependencySettings(bc)
	return nil
}

// Validates config options related to a languages dependencies.
func (lang *Language) validateDependencySettings(bc *ButlerConfig) {
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
		bc.Task.RunAll = true
		lang.DepOptions.ExcludeStdLibs = false
		lang.DepOptions.ExternalDeps = false
		lang.DepCommands.StandardLibrary = ""
		lang.DepCommands.Workspace = ""
		lang.DepCommands.External = ""
	}
}
