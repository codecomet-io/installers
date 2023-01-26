# Bash

Provides a high level API to run bash commands.

## Features

- execution tracing
- err and exit trapping
- debugger hook
- stdout, stderr, exit code capture
- bash configuration exposed

## Gotcha

- requires socat for the debugger to work
- timezone / TZ does not work rn
- if an ERR hook fails
    - if exiting, only the previous error is traced, but this one is the one shown as the final exit code
    - if erroring, we trigger the err handler (again?)
- only tested with Debian for now
- if live debugger has been used and exited, the end debugger does not start (by design)
- /tmp override with a tmpfs is currently at 128M which may not be enough for some commands 

## TODO

- auto connecting debugger would be nice
- remote debugger / ghost tunneling with github auth
- sentry trap reporter on errors would be nice
- provide or mount an editor in the state?
- provide a static mounted bash in a distro less fashion?
- provide a way for the user to use their own .profile / .bashrc / .inputrc?
