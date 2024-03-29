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

There are a handful of options language options that are mandatory for Butler to function. Each of the following options
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

- Description: `filePatterns` is a field for the user to supply file pattern strings associated with code files
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
filePatterns:
  - "package.json"
```
