# Butler

This repository is for the Butler build, test, lint, publish tool. Currently, A reduced version of the new Butler is
being stored here for review. This version is the outer shell for what the generalized Butler will become. This document
will attempt to concisely explain what is contained here.

## Repository Overview

Currently This repository is largely a refactor of how Butler was setting itself up previously. The code to generalize
Butler will not be included until this has been approved. The source code for Butler is contained within the `code`
directory. The `spec` directory is primarily to contain required documentation for the WSU project class this
application is being written for. More directed documentation can be found here in this README. The `sprints` directory
is also a special directory for progress reports for the project class. The neither the `spec` nor `sprints` directory
is required viewing for understanding Butler.

## Implementation steps

At a high level, Butler simply runs commands to lint, test, build, and publish parts of a repository. Here is the end
planned implementation to accomplish this:

1. First, Butler will read from a user defined config called `.butler.base.yaml` which will be defined at the root of
   the repo. This will contain most of the information that will influence Butler's behavior. Butler is also a cli tool,
   so the flags passed via the cli will be merged with the config settings.
2. Next, paths for Butler to search down will be defined in an ignore file called `.butler.ignore.yaml` which will also
   be defined at the root of the repo.

Butler is a CI build tool that collects a set of paths within a repository where linting, testing, building, and
publishing commands can be executed. Butler will determine a unit of code files to execute commands for. This unit will
be referred to as a `workspace`. Butler will determine which workspaces have been changed by checking their git diff. If
a workspace is dirty, then lint, test, and build commands must be run for it. Furthermore, the dependencies of a dirty
workspace will be also be considered dirty and will require a rebuild. If any third party language dependencies have
been changed, workspaces which use that dependency will be considered dirty also. After a set of dirty workspaces has
been acquired, commands will be executed for each.
