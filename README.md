# Butler

This repository is for the Butler build, test, lint, publish tool. Currently, A reduced version of the new Butler is
being stored here for review. This version is the outer shell for what the generalized Butler will become. This document
will attempt to concisely explain what is contained here.

## Butler Config Spec

For information on the Butler config fields, visit the `butler_config.md` file in the spec directory.

## Repository Overview

Currently, what is in this repository is largely a refactor of how Butler was setting itself up previously. The code to generalize
Butler will not be included until this has been approved. The source code for Butler is contained within the `code`
directory. The `spec` directory is primarily to contain required documentation for the WSU project class this
application is being written for. More directed documentation can be found here in this README. The `sprints` directory
is also a special directory for progress reports for the project class. The neither the `spec` nor `sprints` directory
is required viewing for understanding Butler.

### Running Butler

To run butler, execute `butler` in the terminal. Since the current code is the shell for Butler, this will just print
out the Butler config and the list of paths that Butler found. Walking the repo for paths is where the current
implementation of Butler ends.

### Implementation steps

At a high level, Butler simply runs commands to lint, test, build, and publish parts of a repository. Here is the
implementation to accomplish this:

1. First, Butler will read from a user defined config called `.butler.base.yaml` which will be defined at the root of
   the repo. This will contain most of the information that will influence Butler's behavior. Butler is also a cli tool,
   so the flags passed via the cli will be merged with the config settings.

2. Next, paths for Butler to search down will be defined in an ignore file called `.butler.ignore.yaml` which will also
   be defined at the root of the repo. These paths can also be defined in the base config.

3. When the config files have been read from, Butler will then walk through the repository to gather a list of files
   based on the paths the user has allowed and hasn't blocked in the config file.

4. After an array of paths has been collected, Butler will determine units of code files to execute commands for. This
   unit will be referred to as a `workspace`. Workspaces will be constructed based on the filepaths determined
   previously. Workspaces will vary from language to language. For example, a golang workspace is simply a directory
   that contains go files. On the other hand, a workspace may be identified by a single file. For example, a
   package.json file can be used to identify node js/ts workspaces.

5. After workspaces are collected, the dependencies for each workspace will be collected. This includes both internal
   packages and external third-party dependencies.

6. The git diff of the current branch will be taken to obtain a list of directories that have changed files within them.
   Using this list, Butler will determine which workspaces have been changed. These changed workspaces will be marked as
   `dirty`. If a workspace is dirty, all workspaces dependent on that workspace will also be marked as dirty. If any
   third party dependency has been changed, the workspaces using that dependency will be marked dirty.

7. With a set of dirty workspaces for each language, Butler can now create `tasks` for each workspace. Each task is
   added to a queue.

8. Finally, tasks will be executed for each dirty workspace in the repo. The results of each task will be printed to the
   console. Any failed task will cause the entire Butler build the fail, and the log of the failure will be printed.

9. Wether Butler passed or failed, a file called `butler_results.json` will be produced that contains information
   regarding the build. This file will be sent to any subscribers and then stored for future viewing using the Butler
   front end app called `Butler Analytics`.
