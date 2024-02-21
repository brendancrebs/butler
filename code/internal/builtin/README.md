<!--
Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
SEL Confidential
-->

# Built in Butler method

This directory will server as a place for storing language specific methods that are built in to Butler. A directory
will be assigned to each language and will contain all of the supported methods for that language. In the config
directory, the languages.json file will contain a small object for each language. The key will represent the name of
the language and the body will contain information for butler to use. The "aliases" field represents common aliases for
the recognized language name. If Butler cannot find a user supplied language from the list of keys, it will check the
aliases field to attempt to match the language before returning with an error if no language is found. The other fields
contain key to represent the type of method such as a "lint_method" and the value representing the file name where that
method could be found such as "go_lint_method.go" for example. this file value can also be stored as an absolute path to
the method file. When naming a directory for a specific language, make sure that the language name corresponds with the
language key that Butler recognizes. If the names of any files are updated, you must also update the languages.json file
so the file changes are represented there.
