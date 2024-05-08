# Butler

This repository is for the Butler build, test, lint, publish tool.

## Running Butler

To run butler, execute `butler` in the terminal. Run with the `-h, --help` flag for a list of all flags. You may also
edit the `.butler.base.yaml` file to further alter Butler's settings. For information on the Butler config fields, visit
the [butler_config.md][1] file in the spec directory.

## Implementation

At a high level, Butler simply runs commands to lint, test, build, and publish parts of a repository. Here is the
implementation steps to accomplish this:

1. First, Butler will read from a user defined config called `.butler.base.yaml`. This will contain most of the information
   that will influence Butler's behavior. Butler is also a cli tool, so the flags passed via the cli will be merged with
   the config settings. The cli flags will take precedence over the config file settings. The location of this config
   should be passed through the cli with the `-c, --config` flag.

2. Next, paths for Butler to search can be defined in an file called `.butler.paths.yaml`. This file will be defined in
   whatever directory the `.butler.base.yaml` is stored. These paths can also be defined in the base config.

3. When the config files have been read from, Butler will walk through the repository to gather a list of files based on
   the paths the user has allowed and hasn't ignored in the config file.

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

8. The tasks will be executed for each dirty workspace in the repo. The results of each task will be printed to the
   console. Any failed task will cause the entire Butler build the fail, and the log of the failure will be printed.

9. Whether Butler passed or failed, a file called `butler_results.json` will be produced that contains information
   regarding the build. This file will be sent to any subscribers and then stored for future viewing using the Butler
   front end app called `Butler Analytics`.

[1]: ./specs/butler_config.md
