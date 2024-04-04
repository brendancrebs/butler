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

type Language struct {
	Name                      string              `yaml:"name,omitempty"`
	VersionPath               string              `yaml:"versionPath,omitempty"`
	TaskExec                  *TaskCommands       `yaml:"taskCommands,omitempty"`
	DepCommands               *DependencyCommands `yaml:"dependencyCommands,omitempty"`
	Workspaces                []*Workspace        `yaml:"workspaces,omitempty"`
	StdLibDeps                []string            `yaml:"stdLibDeps,omitempty"`
	ExternalDeps              []string            `yaml:"externalDeps,omitempty"`
	FilePatterns              []string            `yaml:"filePatterns,omitempty"`
	BuiltinStdLibsMethod      bool                `yaml:"builtinStdLibsMethod,omitempty"`
	BuiltinWorkspaceDepMethod bool                `yaml:"builtinWorkspaceDepMethod,omitempty"`
	BuiltinExternalDepMethod  bool                `yaml:"builtinExternalDepMethod,omitempty"`
	VersionChanged            bool                `yaml:"versionChanged,omitempty"`
}

type TaskCommands struct {
	SetUp   []string `yaml:"setUp,omitempty"`
	Lint    string   `yaml:"lint,omitempty"`
	Test    string   `yaml:"test,omitempty"`
	Build   string   `yaml:"build,omitempty"`
	Publish string   `yaml:"publish,omitempty"`
}

type DependencyCommands struct {
	StandardLibrary string `yaml:"standardLibrary,omitempty"`
	Workspace       string `yaml:"workspace,omitempty"`
	External        string `yaml:"external,omitempty"`
}

func populateTaskQueue(bc *ButlerConfig, taskQueue *Queue, cmd *cobra.Command) {
	now := time.Now()
	fmt.Fprintf(cmd.OutOrStdout(), "Enumerating repo. Creating build, lint, and test tasks...\n")

	for _, step := range toBuildStep {
		if step >= BuildStepLint && step <= BuildStepPublish {
			for _, lang := range bc.Languages {
				lang.createTasks(taskQueue, step, lang.TaskExec.Lint)
			}
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Done. %s\n\n", time.Since(now).String())
}

// Executes commands that must be run before the creation of tasks
func (lang *Language) preliminaryCommands() (err error) {
	for _, cmd := range lang.TaskExec.SetUp {
		fmt.Printf("\nexecuting: %s...  ", cmd)

		commandParts := splitCommand(cmd)
		if len(commandParts) == 0 {
			fmt.Println("empty command, skipping")
			continue
		}
		cmd := exec.Command(commandParts[0], commandParts[1:]...)

		if output, err := execOutputStub(cmd); err != nil {
			err = fmt.Errorf("error executing '%s'\nerror: %v\noutput: %v", cmd, err, output)
			return err
		} else {
			fmt.Printf("success\n")
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

func (lang *Language) getExternalDeps(bc *ButlerConfig) (err error) {
	if lang.BuiltinStdLibsMethod {
		lang.StdLibDeps, err = builtin.GetStdLibs(lang.Name)
	} else if lang.DepCommands.StandardLibrary != "" {
		lang.StdLibDeps, err = ExecuteUserMethods(lang.DepCommands.StandardLibrary, lang.Name)
	}
	if err != nil {
		return
	}

	if lang.BuiltinExternalDepMethod {
		lang.ExternalDeps, err = builtin.GetExternalDependencies(lang.Name)
	} else {
		lang.ExternalDeps, err = ExecuteUserMethods(lang.DepCommands.External, lang.Name)
	}

	return
}

func (lang *Language) createTasks(taskQueue *Queue, step BuildStep, command string) {
	for _, ws := range lang.Workspaces {
		if ws.IsDirty {
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

func (lang *Language) validateLanguage() error {
	if lang.Name == "" {
		return errors.New("a language supplied in the config without a name. Please supply a language identifier for each language in the config")
	}
	if len(lang.FilePatterns) < 1 {
		return fmt.Errorf("no file patterns supplied for '%s'. Please see the 'FilePatterns' options in the config spec for more information", lang.Name)
	}
	if lang.DepCommands == nil {
		lang.DepCommands = &DependencyCommands{
			StandardLibrary: "",
			Workspace:       "",
			External:        "",
		}
	}
	return nil
}
