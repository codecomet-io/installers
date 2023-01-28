#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail -o monitor

# Fancy prompt adapted from https://github.com/pombadev/fancy-linux-prompt/blob/master/LICENSE under MIT License
__powerline() {
    # Unicode symbols
    readonly GIT_NEED_PULL_SYMBOL='⇣'
    readonly GIT_NEED_PUSH_SYMBOL='⇡'
    readonly PS_SYMBOL='🪐' # 🐧☄'▶⏵▶⏵

    # Solarized colorscheme
    readonly BG_BLUE="\\[$(tput setab 4)\\]"
    readonly BG_COLOR5="\\[\\e[48;5;31m\\]"
    readonly BG_COLOR8="\\[\\e[48;5;161m\\]"
    readonly BG_GREEN="\\[$(tput setab 2)\\]"
    readonly BG_RED="\\[$(tput setab 1)\\]"
    readonly FG_BASE3="\\[$(tput setaf 15)\\]"
    readonly FG_BLUE="\\[$(tput setaf 4)\\]"
    readonly FG_COLOR1="\\[\\e[38;5;250m\\]"
    readonly FG_COLOR6="\\[\\e[38;5;31m\\]"
    readonly FG_COLOR9="\\[\\e[38;5;161m\\]"
    readonly FG_GREEN="\\[$(tput setaf 2)\\]"
    readonly FG_RED="\\[$(tput setaf 1)\\]"
    readonly RESET="\\[$(tput sgr0)\\]"

    __git_info() {
        # no .git directory
    	[ -d .git ] || return 0

        local aheadN
        local behindN
        local branch
        local marks=""
        local stats

        # get current branch name or short SHA1 hash for detached head
        branch="$(git symbolic-ref --short HEAD 2>/dev/null || git describe --tags --always 2>/dev/null)"
        [ -n "$branch" ] || return 0 # git branch not found

        # how many commits local branch is ahead/behind of remote?
        stats="$(git status --porcelain --branch | grep '^##' | grep -o '\[.\+\]$')"
        aheadN="$(echo "$stats" | grep -o 'ahead \d\+' | grep -o '\d\+')"
        behindN="$(echo "$stats" | grep -o 'behind \d\+' | grep -o '\d\+')"
        [ -n "$aheadN" ] && marks+=" $GIT_NEED_PUSH_SYMBOL$aheadN"
        [ -n "$behindN" ] && marks+=" $GIT_NEED_PULL_SYMBOL$behindN"

        # print the git branch segment without a trailing newline
        # branch is modified?
        if [ -n "$(git status --porcelain)" ]; then
            printf "%s" "${BG_COLOR8}▶$RESET$BG_COLOR8 $branch$marks $FG_COLOR9"
        else
            printf "%s" "${BG_BLUE}▶$RESET$BG_BLUE $branch$marks $RESET$FG_BLUE"
        fi
    }


    ps1() {
        # Check the exit code of the previous command and display different
        # colors in the prompt accordingly.
        if [ "$?" -eq 0 ]; then
            local BG_EXIT="$BG_GREEN"
            local FG_EXIT="$FG_GREEN"
        else
            local BG_EXIT="$BG_RED"
            local FG_EXIT="$FG_RED"
        fi

        PS1="$FG_COLOR1$BG_COLOR5 CodeComet \\w "
        PS1+="$RESET${FG_COLOR6}"
        PS1+="$(__git_info)"
        PS1+="$BG_EXIT▶$RESET"
        PS1+="$BG_EXIT$FG_BASE3 ${PS_SYMBOL} ${RESET}${FG_EXIT}▶${RESET} "
    }

    PROMPT_COMMAND=ps1
}

_cc_private::debugger::setup(){
  # Set-up PS1
  cat << EOF > "$CC_TMPFS"/.profile
# export PS1='\[\033[01;32m\]CodeComet\[\033[00m\] \w \$ '
alias l="ls -lA"
source ${BASH_SOURCE[0]}
umask 077
export LSCOLORS=exfxcxdxbxegedabagacad
export CLICOLOR=1

_cc_private::console::inline "$CC_COLOR_GREEN" "$CC_COLOR_BLACK" "Welcome!" printf "You are now debugging your pipeline at the point it stopped.\n"
printf "\n"
#_cc_private::console::inline "$CC_COLOR_BLACK" "$CC_COLOR_BLACK" "         " printf ""
printf "Environment, pwd, and filesystem, are exactly as they were when the last action failed.\n"
printf "You can call the following helpers:\n\n"
_cc_private::console::inline "$CC_COLOR_YELLOW" "$CC_COLOR_BLACK" "ccdebug_stdout" printf "will output your action stdout\n"
_cc_private::console::inline "$CC_COLOR_YELLOW" "$CC_COLOR_BLACK" "ccdebug_stderr" printf "will output your action stderr\n"
_cc_private::console::inline "$CC_COLOR_YELLOW" "$CC_COLOR_BLACK" "ccdebug_stdex " printf "will output your action exit code\n"
_cc_private::console::inline "$CC_COLOR_YELLOW" "$CC_COLOR_BLACK" "ccdebug_action" printf "will output the location of your action script\n\n"
_cc_private::console::inline "$CC_COLOR_GREEN" "$CC_COLOR_BLACK" "Tip           " printf "If you want to re-run your action, just call \\\$(ccdebug_action)\n"

# touch "$CC_TMPFS/.codecomet/connected"
__powerline

EOF

  cat << EOF >> "$CC_TMPFS"/.inputrc
"\e[A": history-search-backward
"\e[B": history-search-forward
set show-all-if-ambiguous on
set completion-ignore-case on

EOF

}

# "on demand
# "Abnormal exit, with non zero grace:
# - start debugger repeatedly, with grace
# - once the other end exits, bail out

# "live", with non zero grace
# - start debugger, with infinity grace
# - if the other end was connected and exited, that is it, end of story
# - if the other end is still running, bring socat forward and wait for the end to exit
# - if there was no other end, we should branch back to the first scenario

# This is called on trap ERR, if a live debugger has been running, and will make a decision to bring it back forward and keep it alive,
# or to drop it
cc::debugger::front(){
  # If we exited successfully, just bail out
  local ex="$1"
  [ "$ex" != 0 ] || return 0
  # If we do not have socat, we do not have a debugger...
  command -v socat >/dev/null || return 0
  cc::logger::debug "[debugger-front] socat is here"
  # If the debugger was disabled, return
  [ "$CC_DEBUGGER_GRACE" != 0 ] || return 0
  cc::logger::debug "[debugger-front] grace is fine"


  # Debugger still live? Bail out if not
  jobs | grep -q -v Done || {
    cc::logger::debug "[debugger-front] no jobs detected. Bailing out."
    return 0
  }

  # Ok, it is live. Do we have a client?
  # XXX unfortunately, this does not work as expected... socat does start the process first, bash sources the files and create the stamp file...
  local lasttry
  local current
  lasttry="$(cat $CC_TMPFS/.codecomet/lasttry 2>/dev/null)" || true
  current="$(date +%s)"
  # The older an unconnected socat could be is 2 seconds - if greater than that, then we definitely have a client
  # However, it is possible that a client has been connected for less than 2 seconds (since socat call) when the failure happens,
  # which means the client will wrongly get the boot... no good solution right now
  [ $(( current - lasttry)) -gt 2 ] && {
  # [ -e "$CC_TMPFS/.codecomet/connected" ] && {
    cc::logger::debug "[debugger-front] we have a connected client"
    # Yes, then foreground and let it sit
    fg 2>/dev/null
  } || {
    cc::logger::debug "[debugger-front] no live client. Kick it out restart."
    # No live client. Kill it, and start again the normal process with timeout and message
    kill %%
    cc::debugger::start "$ex"
  }
}

cc::debugger::live(){
  # If we do not have socat, we do not have a debugger...
  command -v socat >/dev/null || return 0
  # If the debugger was disabled, return
  [ "$CC_DEBUGGER_GRACE" != 0 ] || return 0
  # If the grace was not set, set it now to the default
  [ "$CC_DEBUGGER_GRACE" != "" ] || CC_DEBUGGER_GRACE=20

  _cc_private::debugger::setup
  local lasttry
  local current
  local x=0

  while true; do
    lasttry="$(date +%s)"
    printf "%s" "$lasttry" > "$CC_TMPFS/.codecomet/lasttry"
    HOME="$CC_TMPFS" socat exec:'bash -li',pty,stderr,setsid,sigint,sane tcp:$CC_DEBUGGER_IP:$CC_DEBUGGER_PORT,connect-timeout=1 2>/dev/null && {
      cc::logger::debug "[debugger-live] socat returned successfully, meaning the other end has exited 0"
      break
    } || {
      # Break if we have been hanging out more than a second (meaning the other end did connect, but exit in error)
      current="$(date +%s)"
      if [ $(( current - lasttry)) -gt 1 ]; then
        cc::logger::debug "[debugger-live] socat returned with an error, but after more than 1 second, meaning the other end has exited non zero"
        break
      fi
    }
    x=$(( x + 1 ))
    cc::logger::debug "[debugger-live] sleeping"
    sleep 1
  done
}

cc::debugger::start(){
  # If we exited successfully, just bail out
  local ex="$1"
  [ "$ex" != 0 ] || return 0
  # If we do not have socat, we do not have a debugger...
  command -v socat >/dev/null || return 0
  # If the debugger was disabled, return
  [ "$CC_DEBUGGER_GRACE" != 0 ] || return 0
  # If the grace was not set, set it now to the default
  [ "$CC_DEBUGGER_GRACE" != "" ] || CC_DEBUGGER_GRACE=20

  # Prep-up profile
  _cc_private::debugger::setup
  local lasttry
  local current
  local x=0

  cc::logger::error "Abnormal exit code $ex. If you want to inspect manually, start codecomet-debugger. Otherwise, we will exit after $CC_DEBUGGER_GRACE seconds"
  cc::logger::error "You can also restart the build with CODECOMET_DEBUG=true"

  >&2 _cc_private::console::inline "$CC_COLOR_GREEN" "$CC_COLOR_BLACK" "waiting" printf "Waiting %s seconds for debugger to connect" "$CC_DEBUGGER_GRACE"
  # cc::logger::warning "Waiting "


  # https://medium.com/@JAlblas/tryhackme-what-the-shell-walkthrough-6c0ebe8f854e
  #pty, allocates a pseudoterminal on the target — part of the stabilisation process
  #stderr, makes sure that any error messages get shown in the shell (often a problem with non-interactive shells)
  #sigint, passes any Ctrl + C commands through into the sub-process, allowing us to kill commands inside the shell
  #setsid, creates the process in a new session
  #sane, stabilises the terminal, attempting to “normalise” it.

  while [ "$x" -lt "$CC_DEBUGGER_GRACE" ]; do
    lasttry="$(date +%s)"
    HOME="$CC_TMPFS" socat exec:'bash -li',pty,stderr,setsid,sigint,sane tcp:$CC_DEBUGGER_IP:$CC_DEBUGGER_PORT,connect-timeout=1 2>/dev/null && {
      cc::logger::debug "[debugger-stat] socat returned successfully, meaning the other end has exited 0"
      break
    } || {
      # Timeing out means we wait for a second. Any more than that should mean the connection was succesful
      current="$(date +%s)"
      if [ $(( current - lasttry)) -gt 1 ]; then
        cc::logger::debug "[debugger-stat] socat returned with an error, but after more than 1 second, meaning the other end has exited non zero"
        break
      fi
      >&2 printf "."
    }
    x=$(( x + 1 ))
    sleep 1
  done
}
