<!--
Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
SEL Confidential
-->

# Sprint x Report (8/26/21 - 9/24/2021)

video link: <https://www.youtube.com/watch?v=WikvtXDHAHM&ab_channel=BrendanCrebs>

## What's New (User Facing)

- Base config parsing added
- Executing user defined commands added
- Gathering repo filepaths added

## Work Summary (Developer Facing)

For this sprint I wanted to get off on the right foot so I set up a container filled with configurations to aid in development and uphold code quality standards such linting and testing tools. I developed a base to work from for the rest of the project with the config parsing, file walking, and cli functionality. The config parsing at this stage is very subject to change as there is much I want to add. Much of my time was also spent on design and writing documentation. As the sole developer of this project, that was a challenge and took considerable time.

## Unfinished Work

In this sprint I don't believe I left any work unfinished besides the solutions portion of the final report. As of writing this sprint I did not have the time to finish that portion of the documentation and will get it done ASAP.

## Completed Issues/User Stories

Here are links to the issues that we completed in this sprint:

- [#6](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/6)
- [#4](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/4)
- [#3](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/3)
- [#2](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/2)
- [#1](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/issues/1)

## Incomplete Issues/User Stories

The only thing I would count as incomplete would be that I did not link the issues to the pull request I made before merging it. Also I did not complete the solutions portion of the documentation due to lack of time. I will complete it soon an will assure that this will not happen again.

## Code Files for Review

Please review the following code files, which were actively developed during this sprint, for quality:

- [main.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/main.go)
- [cmd.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/cmd.go)
- [config.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/config.go)
- [create_blt_tasks.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/create_blt_tasks.go)
- [task.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/internal/task.go)
- [common.go](https://github.com/WSUCptSCapstone-F23-S24/sel-lintcli/blob/main/code/helpers/common.go)

## Retrospective Summary

Here's what went well:

- I got a solid base of design documentation completed so I can focus on the coding now.
- I got a solid start to the code to work off of going forward.
- I set up great development environment using a docker container.

Here's what we'd like to improve:

- Making sure to have documentation done on time.
- Keeping better contact with the instructor.

Here are changes we plan to implement in the next sprint:

- I plan to have a working prototype by sprint 2.
- I plan to have a set of language specific methods created for the most major languages.
- I plan to have an updated config design.
