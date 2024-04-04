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

### Mandatory Language options

There are a handful of language options that are mandatory for Butler to function. Each of the following options
must be set for every language defined under the `languages` tag.

#### name

- Type: string

- Description: `name` is the identifier for a language. If you wish to use built in dependency parsing methods for the
  language, the `name` field will need to match one of the supported languages for Butler.

- Example:

```yaml
name: "golang"
```

#### filePatterns

- Type: string array

- Description: `filePatterns` is a field for the user to supply pattern strings associated with code files for
  a language. The file pattern could be a file extension, a common file name, a specific file path, or any combination
  of these. When a directory with a defined file pattern is found, Butler will create a `workspace` for this directory.
  A `workspace` is a directory that contains the relevant files for command execution.

  The commands that get defined for a language will get executed in all `workspaces` that Butler finds for that
  language. Keep this in mind when choosing which file pattern(s) to use. Using the example of Javascript below: a
  `package.json` file will typically live in a parent directory of the actual code files it represents and will contain
  the necessary scripts for building, testing, etc. If you attempt adding the `.js` extension to the list of file
  patterns for Javascript, the commands you give to Butler will be executed in the directory where the `package.json`
  is, AND the directory where the `.js` files are. For a case like this you would select just one of the patterns to
  avoid executing commands in directories they weren't intended to be run in.

- Examples:

Example for Javascript:

```yaml
filePatterns:
  - "package.json"
```

Example for C:

```yaml
filePatterns:
  - ".c"
  - ".h"
```

### Language Commands

For Butler to run lint, test, build, or publish code for a language, appropriate shell commands should be provided
