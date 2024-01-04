<!--
Copyright (c) 2023 Schweitzer Engineering Laboratories, Inc.
SEL Confidential
-->
# Sprint 3 Report (11/5/2023 - 12/5/2023)

video link: <https://youtu.be/qEfhlql4hYI>

## What's New (User Facing)

- Fixed task execution.
- Added nested butler config.
- Added Butler results.

## Work Summary (Developer Facing)

The first priority of this sprint was fixing issues with the sprint2 prototype so that Butler would function properly. This
was soon achieved so a few more features were added. One was functionality for a nested Butler config. After this, a
Butler results feature was added to generate information regarding the build. After these new features were added, the focus
became writing unit tests for all features thus far. As of this report I have achieved 80% code coverage with my unit tests.

## Unfinished Work

In terms of unfinished work, I have not yet achieved 100% code coverage with my unit tests. This is not a huge deal but
it will be a top priority going into next semester to make sure that every pull request and sprint is accompanied by 100%
code coverage. The .butlerignore and improved butler config syntax have been delayed to next sprint due to the other
more pressing issues that were completed this sprint.

## Completed Issues/User Stories

Here are links to the issues completed in this sprint:

- [#23](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/23)
- [#24](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/24)
- [#20](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/20)
- [#19](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/19)
- [#9](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/9)
- [#7](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/7)

## Incomplete Issues/User Stories

- [#12](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/12)<pushed back to sprint 4. See the unfinished
  work section>
  [#25](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/25)<pushed back to sprint 4. See the unfinished
  work section>

## Code Files for Review

Please review the following code files, which were actively developed during this sprint, for quality:

- [jenkinsfile](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/ci/jenkinsfile)
- [main_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/main_test.go)
- [cmd.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/cmd.go)
- [cmd_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/cmd_test.go)
- [config.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/config.go)
- [config_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/config_test.go)
- [create_tasks.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/create_tasks.go)
- [create_tasks_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/create_tasks_test.go)
- [enum_build_status.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/enum_build_status.go)
- [enum_build_status_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/enum_build_status_test.go)
- [enum_build_step.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/enum_build_step.go)
- [enum_build_step_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/enum_build_step_test.go)
- [execute_task.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/execute_task.go)
- [execute_task_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/execute_task_test.go)
- [language_config.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/language_config.go)
- [results.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/results.go)
- [results_test.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/results_test.go)
- [workspaces.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/workspaces.go)
- [go_external_deps_method.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/test_user_methods/get_external_deps/go/go_external_deps_method.go)

## Retrospective Summary

Here's what went well:

- I have largely fixed the errors that were present in the sprint2 prototype.
- I am happy with the status of testing.
- I have a good plan going forward for next semester.

Here's what we'd like to improve:

- By next sprint and for all remaining sprints I want 100% code coverage.
- I need to produce up to date documentation on how to use Butler.
- The Butler config should have a better syntax.

Here are changes we plan to implement in the next sprint:

- 100% code coverage.
- CI on the Bitbucket version of this repo(My internships version)
- Cleaner code and updated methods.
