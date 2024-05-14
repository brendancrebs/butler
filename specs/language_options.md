<!--
Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
SEL Confidential
-->

# Language Options Spec

## Introduction

For Butler to execute `tasks` for a language in your repository, you must first supply information about that language
to the config. You will supply each language in your repo as a list under the label of `languages`. An example
`languages` tag is shown below:

```yaml
languages:
  - name: "golang"
    ...options...
  - name: "python"
    ...options...
  - name: "C#"
    ...options...
```

## Specification

The options described by this spec document will be defined on a per language basis.

### Mandatory Language Options

There are a handful of language options that are mandatory for Butler to function. Each of the following options must be
set for every language defined under the `languages` tag.

#### name

- Type: string

- Description: `name` is the identifier for a language. If you wish to use built in dependency parsing methods for the
  language, the `name` field will need to match one of the supported languages for Butler.

- Config Example:

```yaml
name: "golang"
```

#### workspaceFiles

- Type: string array

- Description: `workspaceFiles` is a field for the user to supply pattern strings associated with code files for a
  language. The file pattern could be a file extension, a common file name, a specific file path, or any combination of
  these. When a directory with a defined file pattern is found, Butler will create a `workspace` for this directory. A
  `workspace` is a directory that contains the relevant files for command execution.

  The commands that get defined for a language will get executed in all `workspaces` that Butler finds for that
  language. Keep this in mind when choosing which file pattern(s) to use. Using the example of Javascript below: a
  `package.json` file will typically live in a parent directory of the actual code files it represents and will contain
  the necessary scripts for building, testing, etc. If you attempt adding the `.js` extension to the list of file
  patterns for Javascript, the commands you give to Butler will be executed in the directory where the `package.json`
  is, AND the directory where the `.js` files are. For a case like this you would select just one of the patterns to
  avoid executing commands in directories they weren't intended to be run in.

  By default, the files that belong to a `workspace` will be every file living in a child directory of the workspace. If
  a child directory satisfies the conditions for a new workspace, it will be considered a distinct workspace and all
  child directories of that workspace will be be associated with it instead of the original higher level workspace. If a
  git diff is detected in a child directory of a workspace, the parent workspace directory will be marked as dirty. If
  you wish to change this behavior, you can use the `filePatterns` option specified below to define specific files to
  associate with a languages workspaces.

- Config Examples:

Example for Javascript:

```yaml
workspaceFiles:
  - "package.json"
```

Example for C:

```yaml
workspaceFiles:
  - ".c"
  - ".h"
```

- Scenarios:

In this example, a config specifies this Golang file pattern for `workspaceFiles`:

```yaml
languages:
  - name: "golang"
    workspaceFiles:
      - ".go"
    taskCommands:
      test: "go test"
```

Here is an example file structure containing code files for a Go project:

```txt
folder-A
  - a.go
  - folder-B
    - b.go
    - c.go
  - folder-C
    - main.ts
```

In this example, a `workspace` will be created for both `folder-A` and `folder-B` due to the `.go` pattern matching the
`a.go`, `b.go` and `c.go` files. A workspace will not be created for `folder-C` since nothing within it matches a
`workspacePattern` supplied in the config. Now, after the workspaces are determined, the `go test` command will be
executed within the `folder-A` and `folder-B` if a git diff indicates they need to be built.

### Optional Language Settings

The remaining settings in this spec are not required for Butler to run. However, it is strongly recommended that a user
reads and understands them so they can properly utilize Butler as designed. In particular, the user should review the
`Language Task Commands` section. Without defining any commands, Butler will not be able to do anything for a language.

#### filePatterns

- Type: string array

- Description: `filePatterns` is a field that is correlated with the `workspaceFiles` setting. For this option a user
  also supplies file patterns that identify relevant files for a language `workspace`. The difference is that a
  workspace will NOT be created if files are found that match the pattern. Instead, if these files are correlated with
  an existing workspace and have a git diff, that workspace will be marked as dirty. A git diff on any file pattern
  listed under `workspaceFiles` will also mark the workspace as dirty. The workspace to be marked dirty will be the
  closest parent workspace directory for the given language.

- Config Example:

Example for Typescript:

```yaml
workspaceFiles:
  - "package.json"
filePatterns:
  - ".ts"
  - ".tsx"
```

### Language Task Commands

For Butler to run `tasks` for a language, appropriate shell commands should be provided for building, testing, linting,
ect. To set any number of these commands, you must first create a `taskCommands` tag in the language options like so:

```yaml
languages:
  - name: "golang"
    taskCommands:
      lint: "example lint command"
      test: "example test command"
    ...options...
  - name: "python"
    ...options...
```

The commands that you supply for a language will be executed in every `workspace` for that language. Each of the
following fields can only be defined under the `taskCommands` tag. All of these commands are optional, but keep in mind
that Butler won't do anything if it isn't provided with commands to execute.

#### setUp

- Type: string array

- Description: `setUp` is a field where you can provide commands that will be executed once as a part of a global set up
  before any `workspaces` are collected, dependencies are collected, or `tasks` executed for a language. You would add
  commands here if you want something about the build server environment or the language to be altered before Butler
  does anything related to that language.

- Config Example:

```yaml
taskCommands:
  setUp:
    - "example preliminary command"
```

#### cleanUp

- Type: string array

- Description: `cleanUp` is similar to `setUp`. The difference is that these commands will be executed after tasks have
  finished. These commands will run regardless of whether the build succeeded or failed.

- Config Example:

```yaml
taskCommands:
  cleanUp:
    - "example teardown command"
```

#### lint

- Type: string

- Description: `lint` is the field where you supply the shell command you wish to have executed during the linting stage
  for the given language.

- Config Example:

```yaml
taskCommands:
  lint: "go lint"
```

#### test

- Type: string

- Description: `test` is the field where you supply the shell command you wish to have executed during the testing stage
  for the given language.

- Config Example:

```yaml
taskCommands:
  lint: "go lint"
  test: "go test"
```

#### build

- Type: string

- Description: `build` is the field where you supply the shell command you wish to have executed during the building
  stage for the given language.

- Config Example:

```yaml
taskCommands:
  lint: "go lint"
  test: "go test"
  build: "go build"
```

#### publish

- Type: string

- Description: `publish` is the field where you supply the shell command you wish to have executed during the publishing
  stage for the given language.

- Config Example:

```yaml
taskCommands:
  lint: "go lint"
  test: "go test"
  build: "go build"
  publish: "go publish"
```

### Language Dependencies Overview

This section relates to the gathering of dependencies for a language. This is completely optional, but the inclusion of
dependency collection can lead to more efficient builds. To enable dependency analysis, you must first set the
`dependencyAnalysis` option under `Dependency Options` to true.

The point of gathering language dependencies is that if certain dependencies are updated in a pull request, all code
using that dependency will need to be built, tested, ect. If you choose not to supply methods for determining a
language's dependencies, Butler cannot determine which code is or isn't using an updated dependency. As a result Butler
will execute a full build every time it's run. This means all possible tasks will run, even on code that doesn't have a
git diff. If Butler is aware of what dependencies have been changed, it can exclude certain code from the build process
which will speed up build times in many cases.

There are three types of dependencies Butler can track. The first are external dependencies which refers to any third
party dependencies that are used by a language in the entire repository. This list of external dependencies will be
added to a list of workspaces which have been marked as `dirty` for having a git diff. This is because those
`workspaces` could be partially or fully exported as dependencies themselves. Therefore they should be treated in the
same manner that updated third party dependencies are.

The second type are called `workspace` dependencies. These are dependencies that are used by a particular `workspace`.
If this feature is utilized, each `workspace` will be tracked by Butler with the list of what dependencies it's using.
For Butler's language dependency feature to function properly, both of these dependency types must be tracked. The
reason is that Butler will attempt to determine which of the external dependencies have been changed. A list of these
changed dependencies will be compared against the dependencies being used in each `workspace` to determine if that
`workspace` is `dirty` or not. A `dirty workspace` is a `workspace` that needs to be included in the Butler build. If a
`workspace` imports something from another workspace that has been marked as dirty, it too will be marked as dirty.

The third type will be standard library dependencies for a language. You may give Butler the option to track these so
that they can be identified and excluded. Standard library dependencies should typically be excluded because we can
assume that any standard library imports are tied to the language version. If the language version has been changed in
the pull request, this will automatically trigger a full Butler build. Otherwise the standard library imports can safely
be excluded from the various `workspace` dependency lists and the external dependency list. If you wish to use Butler's
dependency gathering feature, this is optional but recommended.

### Dependency Options

The following options are settings related to dependency analysis in Butler. However, this is NOT where user supplied
dependency parsing commands will be supplied. To define these options you must create a `dependencyOptions` tag in the
language options like so:

```yaml
languages:
  - name: "golang"
    ...options...
    taskCommands:
      ...options...
    dependencyOptions:
      dependencyAnalysis: true
      excludeStdLibs: true
      externalDependencies: true
  - name: "python"
    ...options...
```

#### dependencyAnalysis

- Type: bool

- Default: false

- Description: `dependencyAnalysis` is an option to set if you want to enable dependency analysis in Butler. All other
  dependency options will depend upon this option being set to `true`. If this option is set to `false`, or not set,
  every Butler build will be a full build.

#### excludeStdLibs

- Type: bool

- Default: false

- Description: `excludeStdLibs` is an option to set if you want to use Butler's built in methods for finding the
  standard library dependencies for a language and removing them from a languages dependency list to improve
  performance. Setting this to true will prioritize Butler's built in method over any user supplied method for this
  task.

#### externalDependencies

- Type: bool

- Default: false

- Description: `externalDependencies` is an option to set if you want to use Butler's built in methods for finding a
  languages third party dependencies. Butler will check if any of these dependencies have changed compared to the main
  branch. A change would include a version change or the addition/removal of a dependency. Butler will mark each
  workspace that used a changed dependency as dirty. Setting this to true will prioritize Butler's built in method over
  any user supplied method for this task.

### Dependency Command Options

The following options relate to user supplied commands that Butler will execute to acquire the dependencies for a
language. To use this feature, you must set the previous `dependencyAnalysis` option to true. These commands should pipe
the dependencies for each task to Butler as a list of strings. The format of these strings will depend on how the
language represents dependencies. If a user defines commands for all of these separate tasks, they should make sure the
dependency string format is consistent between each so the dependencies can be correctly string matched. To define these
options you must create a `dependencyCommands` tag in the language options like so:

```yaml
languages:
  - name: "golang"
    ...options...
    taskCommands:
      ...options...
    dependencyOptions:
      ...options...
    dependencyCommands:
      external: "example command"
  - name: "python"
    ...options...
```

#### standardLibrary

- Type: string

- Description: `standardLibrary` is a command to return an array of standard library dependencies for a language. This
  command should return a list of dependencies represented by a list of strings. Like the other options, this list
  should be piped to butler. Optionally, the user can supply whether the language version corresponding with the
  standard library dependencies has changed by returning a boolean `true` or `false` as the first value of the list. If
  a language version has changed, a full build will be triggered.

- Config Example:

```yaml
dependencyCommands:
  standardLibrary: "example command"
```

#### workspace

- Type: string

- Description: `workspace` is a command to return an array of dependencies for a particular workspace. This command
  should simply return the dependencies for a single workspace. Expect that this command will be executed at the
  location of every workspace that was collected for the given language.

  When a particular workspace imports other workspaces as dependencies, those dependencies have to be matched with the
  workspaces Butler has collected independently. Because of this, the internal dependencies should be formatted in one
  of two ways so Butler can properly match the path strings. One option is a path to the workspace extending from the
  `workspaceRoot` defined in the config. The next would be to supply the absolute path to the workspace. For clarity,
  paths to workspaces are represented as absolute paths within Butler.

  External dependencies however will be represented in the specified format for that language. The list of dependency
  string formats can be found in the [language_dependencies.md][1] spec.

- Config Example:

```yaml
dependencyCommands:
  workspace: "example command"
```

#### external

- Type: string

- Description: `external` is a command to return an array of external third party dependencies for a language that have
  been changed. The string format to represent these dependencies will depend on the language.

- Config Example:

```yaml
dependencyCommands:
  external: "example command"
```

[1]: ./language_dependencies.md
