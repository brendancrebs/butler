<!--
Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
SEL Confidential
-->

# Language Format Spec

This spec document is a list of dependency information Butler. Each Language in this spec will be supplied with
its required language ID, the format of its dependencies, and any additional information that may be relevant for
getting Butler to handle it correctly.

## Language ID

The `language ID` relates to the `Name` field specified in the language_options.md spec. This required field will be the
display name for tasks while they're being executed. For languages that Butler supports for dependency analysis, the
specified language ID is relevant so Butler knows which methods to use. This is because the variance between how
languages track dependencies is significant enough that it cannot be generalized while preserving accuracy. If you wish
to use Butler's built in dependency analysis, the name field for your language must match the corresponding language ID
listed in this file.

## Dependency Format

### External Dependencies

An `external dependency` is any dependency whose source is external to the repository.

User made dependency analysis methods are expected to return a list of strings that represent dependencies. The
`dependency format` for a language will specify the format for how external dependencies should be represented in string
form. A common format for language dependencies is important because Butler will preform string comparisons between the
dependencies of an individual `workspace` against a list of dependencies that have changed. If a dependency is listed in
two separate formats, Butler might not be able to detect that a changed dependency was present in a particular
workspace. It is only relevant to follow the suggested external dependency format if the user does not also supply a
`workspace` dependency method. In that case they may use their own defined format since Butler will not be parsing
dependencies on its own.

### Internal Dependencies

An `internal dependency` is a dependency that lives within the repository Butler runs in.

Internal dependencies will not follow a language specific format like external dependencies do. Instead, internal
dependencies will be represented by an absolute path to a specific workspace. If a workspace imports code from another
workspace, an absolute path to the workspace being imported should be returned as a dependency of the workspace
importing it.

## Golang

- Language ID: `golang`

- Dependency Format: For `golang`, an external dependency should be represented as it is found in the `go.mod` file.
  The version should not be included since only dependencies that have changed should be returned. An example list of go
  dependencies might look like, `github.com/smartystreets/goconvey, github.com/spf13/cobra, gopkg.in/yaml.v2`.
