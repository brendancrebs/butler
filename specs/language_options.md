<!--
Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
SEL Confidential
-->

# Language Options Spec

## Introduction

For Butler to execute tasks for a language in your repository, you must first supply information about that language to
the config. You will supply each language for Butler to recognize as a list under the label of `languages`. An example is
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

## Specification

### General Language options

Below are the options for each individual language that gets defined under the `languages` tag.

#### name

- Type: string

- Required: Yes

- Description: `name` is the identifier for a language. This is a mandatory option. If you wish to use built in dependency parsing methods for the
  language, the `name` field will need to match one of the supported languages for Butler.

- Example:

```yaml
name: "golang"
```

#### filePatterns

- Type: string array

- Required: Yes

- Description: `filePatterns` is a required field for the user to supply file pattern strings associated with code files
  for this language. Butler will use this to build workspaces. The file pattern could be a file extension, a common file
  name, or a specific file path. Keep in mind that Butler will execute the commands you supply in the directories it
  find these files.

- Examples:

Example for C:

```yaml
filePatterns:
  - ".c"
  - ".h"
```

Example for Javascript:

```yaml
FilePatterns:
  - "package.json"
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
      lint: "example lint command"
      test: "example test command"
  - name: "python"
    ...options...
```

#### setUp

- Type: string array

- Description: `setUp` is a list of commands that would need to be executed before the execution of tasks or
  gathering of dependencies for a language.

- Example

```yaml
taskCommands:
  setUp:
    - "example preliminary command"
```

#### lint

- Type: string

- Description: `lint` is the field where you supply the command you wish to have executed during the linting
  stage for the given language.

- Example:

```yaml
taskCommands:
  lint: "go lint"
```

#### test

- Type: string

- Description: `test` is the field where you supply the command you wish to have executed during the testing
  stage for the given language.

- Example:

```yaml
taskCommands:
  lint: "go lint"
  test: "go test"
```

#### build

- Type: string

- Description: `build` is the field where you supply the command you wish to have executed during the building
  stage for the given language.

- Example:

```yaml
taskCommands:
  lint: "go lint"
  test: "go test"
  build: "go build"
```

#### publish

- Type: string

- Description: `publish` is the field where you supply the command you wish to have executed during the publishing
  stage for the given language.

- Example:

```yaml
taskCommands:
  lint: "go lint"
  test: "go test"
  build: "go build"
  publish: "go publish"
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
      external: "example command"
  - name: "python"
    ...options...
```

#### standardLibrary

- Type: string

- Description: `standardLibrary` is a command to return an array of standard library dependencies for a language.

- Example:

```yaml
dependencyCommands:
  standardLibrary: "example command"
```

#### workspace

- Type: string

- Description: `workspace` is a command to return an array of dependencies for a particular workspace. Expect
  that this command will be executed at the location of every workspace that was collected for the given language.

- Example:

```yaml
dependencyCommands:
  workspace: "example command"
```

#### external

- Type: string

- Description: `external` is a command to return an array of external third party dependencies for a language.

- Example:

```yaml
dependencyCommands:
  external: "example command"
```

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
