#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail

echo "DO NOT USE THIS"
exit

readonly _HERE=$(cd "$(dirname "${BASH_SOURCE[0]:-$PWD}")" 2>/dev/null 1>&2 && pwd)
readonly _BASE="$HOME/.codecomet/install"
readonly _PRIVATE="${PRIVATE:-false}"
readonly _RELEASE=mark-I

############################
# generic Helpers
############################
COLOR_RED=1
COLOR_GREEN=2
COLOR_YELLOW=3

# Prefix a date to a log line and output to stderr
_stamp(){
  local color="$1"
  local level="$2"
  local i
  shift
  shift
  [ "$TERM" ] && [ -t 2 ] && >&2 tput setaf "$color"
  for i in "$@"; do
    >&2 printf "[%s] [%s] %s\\n" "$(date)" "$level" "$i"
  done
  [ "$TERM" ] && [ -t 2 ] && >&2 tput op
}

logger_info(){
  _stamp "$COLOR_GREEN" "INFO" "$@"
}

logger_warning(){
  _stamp "$COLOR_YELLOW" "WARNING" "$@"
}

logger_error(){
  _stamp "$COLOR_RED" "ERROR" "$@"
}

ensure::system(){
  mkdir -p "$_BASE"

  # Enforce XCode command line tools to be here
  while ! git --help 1>/dev/null 2>&1; do
    logger_warning "You must install git. This is typically provided by XCode command line tools. You should be prompted to install it now."
    logger_info "Hit enter when XCode installation completes, or manually install git yourself and relaunch this script making sure git is in the PATH."
    read -r
  done
}

ensure::self(){
  [ -e "$_BASE/repo" ] || {
    git clone https://github.com/codecomet-io/installers "$_BASE/repo"
  }
}

ensure::brew(){
  [ "$_PRIVATE" == true ] || return 0
  [ -e "$_BASE/brew" ] || {
    git clone https://github.com/Homebrew/brew "$_BASE/brew"
    config::brew
    update::brew
  }
}

config::brew(){
  [ "$_PRIVATE" == true ] || return 0
  # Reset env, just in case
  while read -r line; do
    unset "${line%%=*}"
  done < <(env | grep HOMEBREW)

  # Source env
  eval "$($_BASE/brew/bin/brew shellenv)"
}

update::brew(){
  brew update --force --quiet
}

update::repo(){
  cd "$_BASE/repo" || exit
  logger_error "XXX disabled for the time being because of local development"
  # git pull --rebase --quiet
  cd - > /dev/null || exit
}

update::deps(){
  brew upgrade
}

install::deps(){
  brew install qemu
}

install::cc(){
  "$_HERE"/release/"$_RELEASE"/bin/codecomet-machine install
}

uninstall::cc(){
  "$_HERE"/release/"$_RELEASE"/bin/codecomet-machine uninstall
}


# If we are not the installed version...
[ "$0" == "$_BASE/repo/install.sh" ] || {
  # Check system req
  ensure::system
  logger_info "System requirements ok"

  # Ensure we have a private brew
  ensure::brew
  logger_info "Private brew installed"

  # Ensure we have a checkout of ourselves
  ensure::self
  logger_info "Initial checkout done"

  # Call ourselves
  config::brew
  "$_BASE/repo/install.sh" "$@"
  exit
}


# Now, dependent on command
case "${1:-}" in
"update")
  logger_info "Updating codecomet"
  uninstall::cc

  logger_info "Updating brew dependencies"
  update::brew
  logger_info "Updating installation repo"
  update::repo
  logger_info "Updating dependencies"
  update::deps

  install::cc
  ;;
"install"|"")
  logger_info "Installation"
  install::deps
  install::cc
  ;;
*)
    logger_error "Unknown command: $1"
    exit 1
    ;;
esac

