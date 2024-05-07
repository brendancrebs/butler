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
`.butler.ignore.yaml` and will contain paths within a repo that should be either allowed or ignored by Butler.

## Specification

### Mandatory Options

The following are mandatory options for the `.base.butler.yaml`. The header for each option is the key for that option which should
be used in the file.

#### workspaceRoot

- Type: string

- Description: `workspaceRoot` is a path that will specify where the root of the repository is in the local
  filesystem. Butler can only observe paths that are children of the `workspaceRoot` path. Filepaths supplied in other
  options should only be children of the workspace root. This is the only required field in the base Butler config options.

- Example:

```yaml
workspaceRoot: "/workspaces/butler"
```

### General Options

The following fields are optional. It's still highly recommended for a user to read through these options to configure
Butler to best suit their needs.

#### allowedPaths

- Type: string array

- Description: The `allowedPaths` option is a list of folders that `Butler` is permitted to search in order to build
  tasks. If any paths are supplied to this option, `Butler` will ONLY search down these paths for `workspaces`. For
  information on `workspaces`, see the `workspaces` section of the glossary at the bottom of this document.

  If the `allowedPaths` option is not initialized, Butler will search every path in a repo that isn't explicitly
  ignored. Any directory the user supplies will be recursively searched for more `workspaces`.

- Example:

```yaml
allowedPaths:
  - apps/butler
  - interfaces
  - lib/helpers
```

#### ignorePaths

- Type: string array

- Description: The `ignorePaths` option is similar to the `allowedPaths` option. `ignorePaths` will instead contain a
  list of folder paths that Butler should not look down. It should be noted that if an ignored path is the child of an
  allowed path, the ignored path will still be ignored. It is recommended that a user includes generated directories
  such as node's `node_modules` in this list. Large directories that don't have relevant code will increase build times
  and cause unintended `workspace` collection.

- Example:

```yaml
ignore-paths:
  - node_modules
  - apps/butler/test_data
  - scripts
```

#### criticalPaths

- Type: string array

- Description: `criticalPaths` is a list of paths which should trigger a `full build` if any of them have been changed.
  These paths can either be file or directory paths. For directory paths, if any file has been changed in that
  directory, a full build will be triggered. It is recommended that you add the `.butler.base.yaml` file location to this
  list so that a full build is ran if `Butler's` behavior is altered via the config.

- Example:

```yaml
critical-paths:
  - lib/interfaces
  - apps/example/critical
  - .butler.base.yaml
```

#### publishBranch

- Type: string

- Description: The `publishBranch` option represent the main development branch for a repository. This will also
  represent the branch Butler will publish from. This option is not mandatory. However, If this option is set, `Butler`
  will diff the current branch against the `publishBranch` to determine what needs to be built based on what has been
  changed from the main branch. If this option is not set, a git diff will not be taken and a full build will be
  triggered every time.

- Example:

```yaml
publishBranch: "main"
```

#### resultsFilePath

- Type: string

- Description: `resultsFilePath` is the path/filename for where a butler results json file should be
  generated. This file name must also be added to the repository `.gitignore` file.

- Example:

```yaml
resultsFilePath: "./butler_results.json"
```

#### runAll

- Type: bool

- Default: false

- Description: `runAll` enabled a full build. This means all tasks should be run regardless of the git diff. A full
  build will also be triggered if any of the following conditions are met:
  1. The a `publishBranch` variable is not set. This can be set through the `publishBranch` config option or the
     `--publish-branch` cli flag.
  2. The environment variable `BUTLER_SHOULD_RUN_ALL` is set to true.
  3. The git branch Butler runs on is the `publish branch`.
  4. A `critical path` has been changed.
  5. `Dependency analysis` is disabled.
  6. The version of a language has been changed.

#### lint

- Type: bool

- Default: false

- Description: `lint` enables linting tasks. NOTE: if this is set to false a full build will still not execute lint
  tasks.

#### test

- Type: bool

- Default: false

- Description: `test` enables testing tasks.

#### build

- Type: bool

- Default: false

- Description: `build` enables building tasks.

#### publish

- Type: bool

- Default: false

- Description: `publish` enables publishing tasks.

### Language Options

IMPORTANT: Butler requires information about the languages in a repository. For detailed information about what
information to supply and how, see the [language_options.md][1] spec file.

### Butler Ignore

The `.butler.ignore.yaml` file can be used to store the `allowedPaths` and `ignoredPaths`. To use this feature, add a
file with the name `.butler.ignore.yaml` at the root of the repo. Then you may add the allowed/ignored paths with this
same syntax as specified for the allowed/ignored paths in the base config above.

#### Butler Ignore Example

```yaml
allowedPaths:
  - apps/butler
  - interfaces
  - lib/helpers

ignoredPaths:
  - node_modules
  - apps/butler/test_data
  - scripts
```

## Glossary

`Workspaces`: A `workspace` is a directory that contains the code files needed for `task` execution. For each language,
file patterns will be defined to Butler to identify a language's `workspaces`. For more information on defining
languages to Butler, see the [language_options][1] spec.

`tasks`: A `task` is a process that Butler will run in workspace. `Tasks` should be commands that can be executed in a
shell. These tasks should fulfill a certain purpose such as building, or testing code. If you have enabled building and
testing for a language, Butler will create a task for building, and a task for testing in every `workspace` for that
language. `Task` commands should either pass or fail. Tasks that fail should print their logs for Butler to display.

`full build`: A `full build` refers to a Butler build where every task is run regardless of the git diff. This event can
be triggered by a number of conditions such as: the repo not using git, full builds explicitly defined in the config, a
`criticalPath` having a git diff, or the environment variable, `BUTLER_SHOULD_RUN_ALL` being set to true.

[1]: ./language_options.md
