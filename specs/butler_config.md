# Butler Config Spec

<!--
Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
SEL Confidential
-->

## Introduction

The `Butler Config` is the file that will determine the majority of the `Butler's` behavior. This spec will explain the
various options the Butler config and cli offer and example usage of each. The config is a yaml file and is named
`.butler.base.yaml` by default. A yaml file of a different name can be specified as the `Butler Config` if the path to it is
passed through the cli. Beyond the base config, a `Butler ignore` file can also be specified. This file must be named
`.butler.ignore.yaml` and will contain paths within a repo that should be either allowed or blocked by Butler.

## Specification

### General Config Options

The following are options for the `.base.butler.yaml`. The header for each option is the key for that option which should
be used in the file.

#### allowed-paths

- Type: string array

- Description: The `allowed-paths` option is a list of file paths/patterns that `Butler` is permitted to look down in order to build tasks.
  `Butler` will only look down paths that are included in the `allowed-paths` field.

- Example:

```yaml
allowed-paths:
  - apps/butler
  - interfaces
  - lib/helpers
```

#### blocked-paths

- Type: string array

- Description: The `blocked-paths` option is similar to the `allowed-paths` option. `Blocked-paths` will instead contain a list of
  filepaths that Butler should not look down.

- Example:

```yaml
blocked-paths:
  - node_modules
  - apps/butler/test_data
  - scripts
```

#### critical-paths

- Type: string array

- Description: `critical-paths` is a list of paths which should trigger a full Butler build if any of them have
  been changed. These paths can either be file or directory paths. For directory paths, if any file has been changed in
  that directory, a full build will be triggered. It is recommended that you add the `.butler.base.yaml` file location to this
  list so that a full build is ran if `Butler's` behavior is altered.

- Example:

```yaml
critical-paths:
  - lib/interfaces
  - apps/example/critical
  - .butler.base.yaml
```

#### publish-branch

- Type: string

- Description: The `Publish-branch` option represent the main development branch for a repository. This will also
  represent the branch Butler will publish from.

- Example:

```yaml
publish-branch: "main"
```

#### Results-file-path

- Type: string

- Description: `results-file-path` is the path/filename for where a butler results json file should be
  generated. This file name must also be added to the repository `.gitignore` file.

- Example:

```yaml
results-file-path: "./butler_results.json"
```

#### workspace-root

- Type: string

- Description: `workspace-root` is a path that will specify where the root of the repository is in the local
  filesystem. Butler can only observe paths that are children of the `workspace-root` path. Filepaths supplied in other
  fields should only be children of the workspace root.

- Example:

```yaml
workspace-root: "/workspaces/butler"
```

#### git-repository

- Type: bool

- Default: false

- Description: `git-repository` specifies if Butler is working within a git repository. If this option is true,
  `Butler` will diff the current branch against the `publish-branch` to determine what needs to be built based on what has
  been changed from the main branch. If this option is set to false, a full build will be triggered every time.

#### should-run-all

- Type: bool

- Default: false

- Description: `should-run-all` enabled a full build. This means all tasks should be run regardless of the git diff.

#### should-lint

- Type: bool

- Default: false

- Description: `should-lint` enables linting tasks. NOTE: if this is set to false a full build will still not execute lint
  tasks.

#### should-test

- Type: bool

- Default: false

- Description: `should-test` enables testing tasks.

#### should-build

- Type: bool

- Default: false

- Description: `should-build` enables building tasks.

#### should-publish

- Type: bool

- Default: false

- Description: `should-publish` enables publishing tasks.

### Language Options

For Butler to execute tasks for a language in your repository you must first supply information about that language to
the config. You will supply each language you want Butler to recognize as a list under the label of `languages`. An example is
shown below:

```yaml
languages:
  - name: "golang"
    ...options...
  - name: "python"
    ...options...
  - name: "C#"
    ...options...
```

Below are the options you may set for each individual language you define under the `languages` tag.

#### name

- Type: string

- Description: `name` is the identifier for a language. This is a mandatory option. If you wish to use built in dependency parsing methods for the
  language, the `name` field will need to match one of the supported languages for Butler.

- Example:

```yaml
name: "golang"
```

#### fileExtension

- Type: string

- Description: `fileExtension` is an optional field to supply a file extension string associated with code files for
  this language. Butler will use this to build workspaces. If this setting is not defined then the workspaceFile setting
  must be set instead.

- Example:

```yaml
fileExtension: ".go"
```

#### workspaceFile

- Type: string

- Description: `workspaceFile` is an optional field to supply a file pattern that Butler will use to establish a
  workspace location for the associated language. If this setting is not defined then you must set the fileExtension
  setting instead.

- Example:

```yaml
workspaceFile: "package.json"
```

#### BuiltinStdLibsMethod

- Type: bool

- Default: false

- Description: `BuiltinStdLibsMethod` is an option to define if you want to use Butler's built in methods for
  determining standard library dependencies for a language.

#### BuiltinWorkspaceDepMethod

- Type: bool

- Default: false

- Description: `BuiltinWorkspaceDepMethod` is an option to define if you want to use Butler's built in methods for
  determining the dependencies used for each workspace.

#### BuiltinExternalDepMethod

- Type: bool

- Default: false

- Description: `BuiltinExternalDepMethod` is an option to define if you want to use Butler's built in methods for
  determining the external dependencies for the given language.

### Task Command Options

The following options relate to commands that Butler will execute for each stage of a languages build process. To define
these options you must create a `taskCommands` tag in the language options like so:

```yaml
languages:
  - name: "golang"
    ...options...
    taskCommands:
      lintCommand: "example lint command"
      testCommand: "example test command"
  - name: "python"
    ...options...
```

#### setUpCommands

- Type: string array

- Description: `setUpCommands` is a list of commands that would need to be executed before the execution of tasks or
  gathering of dependencies for a language.

- Example

```yaml
taskCommands:
  setUpCommands:
    - "example preliminary command"
```

#### lintCommand

- Type: string

- Description: `lintCommand` is the field where you supply the command you wish to have executed during the linting
  stage for the given language.

- Example:

```yaml
taskCommands:
  lintCommand: "go lint"
```

#### testCommand

- Type: string

- Description: `testCommand` is the field where you supply the command you wish to have executed during the testing
  stage for the given language.

- Example:

```yaml
taskCommands:
  lintCommand: "go lint"
  testCommand: "go test"
```

#### buildCommand

- Type: string

- Description: `buildCommand` is the field where you supply the command you wish to have executed during the building
  stage for the given language.

- Example:

```yaml
taskCommands:
  lintCommand: "go lint"
  testCommand: "go test"
  buildCommand: "go build"
```

#### publishCommand

- Type: string

- Description: `publishCommand` is the field where you supply the command you wish to have executed during the publishing
  stage for the given language.

- Example:

```yaml
taskCommands:
  lintCommand: "go lint"
  testCommand: "go test"
  buildCommand: "go build"
  publishCommand: "go publish"
```

### Dependency Command options

The following options relate to commands that Butler will execute to acquire the dependencies for a language. To define
these options you must create a `dependencyCommands` tag in the language options like so:

```yaml
languages:
  - name: "golang"
    ...options...
    taskCommands:
      ...options...
    dependencyCommands:
      externalDepCommand: "example command"
  - name: "python"
    ...options...
```

#### StdLibsCommand

- Type: string

- Description: `StdLibsCommand` is a command to return an array of standard library dependencies for a language.

- Example:

```yaml
dependencyCommands:
  StdLibsCommand: "example command"
```

#### internalDepCommand

- Type: string

- Description: `internalDepCommand` is a command to return an array of dependencies for a particular workspace. Expect
  that this command will be executed at the location of every workspace that was collected for the given language.

- Example:

```yaml
dependencyCommands:
  internalDepCommand: "example command"
```

#### externalDepCommand

- Type: string

- Description: `externalDepCommand` is a command to return an array of external third party dependencies for a language.

- Example:

```yaml
dependencyCommands:
  externalDepCommand: "example command"
```

### Butler Ignore

The `.butler.ignore.yaml` file can be used to store the `allowed-paths` and `blocked-paths`. To use this feature, add a
file with the name `.butler.ignore.yaml` at the root of the repo. Then you may add the allowed/blocked paths with this
same syntax as specified for the allowed/blocked paths in the base config above.

#### Butler Ignore Example

```yaml
allowed-paths:
  - apps/butler
  - interfaces
  - lib/helpers

blocked-paths:
  - node_modules
  - apps/butler/test_data
  - scripts
```
