package bash

import (
	"reflect"
	"strings"
)

type Config struct {
	// BashPath []string

	// -a Automatically mark variables and functions which are modified or created for export to the environment of
	// subsequent commands.
	AllExport bool `codecomet:"allexport"`

	// -B The shell performs brace expansion (see Brace Expansion above). This is on by default.
	BraceExpand bool `codecomet:"braceexpand"`

	// Use an emacs-style command line editing interface. This is enabled by default when the shell is interactive, unless the shell is started with the --noediting option. This also affects the editing interface used for read -e.
	Emacs bool `codecomet:"emacs"`

	// -e Exit immediately if a pipeline (which may consist of a single simple command), a subshell command enclosed
	// in parentheses, or one of the commands executed as part of a command list enclosed by braces (see SHELL GRAMMAR
	// above) exits with a non-zero status. The shell does not exit if the command that fails is part of the command
	// list immediately following a while or until keyword, part of the test following the if or elif reserved words,
	// part of any command executed in a && or â”‚â”‚ list except the command following the final && or â”‚â”‚, any
	// command in a pipeline but the last, or if the command's return value is being inverted with !. A trap on ERR, if
	// set, is executed before the shell exits. This option applies to the shell environment and each subshell
	// environment separately (see COMMAND EXECUTION ENVIRONMENT above), and may cause subshells to exit before
	// executing all the commands in the subshell.
	ErrExit bool `codecomet:"errexit"`

	// -E If set, any trap on ERR is inherited by shell functions, command substitutions, and commands executed in a
	// subshell environment. The ERR trap is normally not inherited in such cases.
	ErrTrace bool `codecomet:"errtrace"`

	// -T If set, any traps on DEBUG and RETURN are inherited by shell functions, command substitutions, and commands
	// executed in a subshell environment. The DEBUG and RETURN traps are normally not inherited in such cases.
	FuncTrace bool `codecomet:"functrace"`

	// -h Remember the location of commands as they are looked up for execution. This is enabled by default.
	Hashall bool `codecomet:"hashall"`

	// -H Enable ! style history substitution. This option is on by default when the shell is interactive.
	HistExpand bool `codecomet:"histexpand"`

	// Enable command history, as described above under HISTORY. This option is on by default in interactive shells.
	History bool `codecomet:"history"`

	// The effect is as if the shell command ''IGNOREEOF=10'' had been executed (see Shell Variables above).
	IgnoreEOF bool `codecomet:"ignoreeof"`

	// If set, allow a word beginning with # to cause that word and all remaining characters on that line to be ignored
	// in an interactive shell (see COMMENTS above). This option is enabled by default.
	InteractiveComments bool `codecomet:"interactive-comments"`

	// -k All arguments in the form of assignment statements are placed in the environment for a command, not
	// just those that precede the command name.
	Keyword bool `codecomet:"keyword"`

	// -m Monitor mode. Job control is enabled. This option is on by default for interactive shells on systems that
	// support it (see JOB CONTROL above). Background processes run in a separate process group and a line containing
	// their exit status is printed upon their completion.
	Monitor bool `codecomet:"monitor"`

	// -C If set, bash does not overwrite an existing file with the >, >&, and <> redirection operators.
	// This may be overridden when creating output files by using the redirection operator >| instead of >.
	NoClobber bool `codecomet:"noclobber"`

	// -n Read commands but do not execute them. This may be used to check a shell script for syntax errors.
	// This is ignored by interactive shells.
	NoExec bool `codecomet:"noexec"`

	// -f Disable pathname expansion.
	NoGlob bool `codecomet:"noglob"`

	// ? ignored?
	NoLog bool `codecomet:"nolog"`

	// -b Report the status of terminated background jobs immediately, rather than before the next primary prompt. This
	// is effective only when job control is enabled.
	Notify bool `codecomet:"notify"`

	// -u Treat unset variables and parameters other than the special parameters "@" and "*" as an error when performing
	// parameter expansion. If expansion is attempted on an unset variable or parameter, the shell prints an error message,
	// and, if not interactive, exits with a non-zero status.
	NoUnset bool `codecomet:"nounset"`

	// -t Exit after reading and executing one command.
	OneCmd bool `codecomet:"onecmd"`

	// -P If set, the shell does not follow symbolic links when executing commands such as cd that change the current
	// working directory. It uses the physical directory structure instead. By default, bash follows the logical chain
	// of directories when performing commands which change the current directory.
	Physical bool `codecomet:"physical"`

	// If set, the return value of a pipeline is the value of the last (rightmost) command to exit with a non-zero status, or zero if all commands in the pipeline exit successfully. This option is disabled by default.
	PipeFail bool `codecomet:"pipefail"`

	// Change the behavior of bash where the default operation differs from the POSIX standard to match the standard (posix mode).
	Posix bool `codecomet:"posix"`

	// EXPERT ONLY
	// -p Turn on privileged mode. In this mode, the $ENV and $BASH_ENV files are not processed, shell functions are not
	// inherited from the environment, and the SHELLOPTS, BASHOPTS, CDPATH, and GLOBIGNORE variables, if they appear in
	// the environment, are ignored. If the shell is started with the effective user (group) id not equal to the real user
	// (group) id, and the -p option is not supplied, these actions are taken and the effective user id is set to the real
	// user id. If the -p option is supplied at startup, the effective user id is not reset. Turning this option off causes
	// the effective user and group ids to be set to the real user and group ids.
	Privileged bool `codecomet:"privileged"`

	// -v Print shell input lines as they are read.
	Verbose bool `codecomet:"verbose"`

	// Use a vi-style command line editing interface. This also affects the editing interface used for read -e.
	// Vi bool `codecomet:"vi"`

	// -x After expanding each simple command, for command, case command, select command, or arithmetic for command,
	// display the expanded value of PS4, followed by the command and its expanded arguments or associated word list.
	XTrace bool `codecomet:"xtrace"`
}

func (cnf *Config) toString() string {
	com := []string{
		"set",
	}

	val := reflect.ValueOf(*cnf)
	typeOf := val.Type()

	for i := 0; i < val.NumField(); i++ {
		serial := typeOf.Field(i).Tag.Get("codecomet")
		if serial == "" {
			serial = strings.ToLower(typeOf.Field(i).Name)
		}

		if val.Field(i).Interface().(bool) {
			com = append(com, "-o")
		} else {
			com = append(com, "+o")
		}
		com = append(com, serial)
	}
	return strings.Join(com, " ")
}
