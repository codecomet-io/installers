#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail -o monitor

readonly CC_COLOR_BLACK=0
readonly CC_COLOR_RED=1
readonly CC_COLOR_GREEN=2
readonly CC_COLOR_YELLOW=3
readonly CC_COLOR_BLUE=4
readonly CC_COLOR_MAGENTA=5
readonly CC_COLOR_CYAN=6
readonly CC_COLOR_WHITE=7

readonly CC_COLOR_DEFAULT_FRONT="$CC_COLOR_WHITE"
readonly CC_COLOR_DEFAULT_BACK="$CC_COLOR_BLACK" #BLUE"

readonly CC_LOGGER_DEBUG=4
readonly CC_LOGGER_INFO=3
readonly CC_LOGGER_WARNING=2
readonly CC_LOGGER_ERROR=1

_cc_private::tput(){
  # Could try harder to figure out where we are at
  # https://stackoverflow.com/questions/911168/how-can-i-detect-if-my-shell-script-is-running-through-a-pipe
  # Prior was [ ! -t 2 ]
  # Could be: test -t 0 (formerly tty -s)
  # About NO_COLOR: https://no-color.org/ with NO_COLOR
  [ ! "$TERM" ] || [ "$NO_COLOR" != "" ] || tput "$@" 2>/dev/null || true
}

_cc_private::console(){
  local front="$1"
  local back="$2"
  local label="$3"
  shift
  shift
  shift

  [ "$label" ] && {
    _cc_private::tput setaf "$back"
    _cc_private::tput setab "$front"
    printf " %-7s " "$label"
    _cc_private::tput setaf "$front"
    _cc_private::tput setab "$CC_COLOR_DEFAULT_BACK"
    printf "▶ "
  } || {
    [ ! "$front" ] || _cc_private::tput setaf "$front"
    [ ! "$back" ] || _cc_private::tput setaf "$back"
  }

  "$@"

  cc::console::reset
}

_cc_private::console::inline(){
  local front="$1"
  local back="$2"
  local label="$3"
  shift
  shift
  shift

  [ "$label" ] && {
    _cc_private::tput setaf "$back"
    _cc_private::tput setab "$front"
    printf " %-s " "$label"
    _cc_private::tput setaf "$front"
    _cc_private::tput setab "$CC_COLOR_DEFAULT_BACK"
    printf "▶ "
  } || {
    [ ! "$front" ] || _cc_private::tput setaf "$front"
    [ ! "$back" ] || _cc_private::tput setaf "$back"
  }

  "$@"
}

cc::console::error(){
  _cc_private::console "$CC_COLOR_RED" "$CC_COLOR_WHITE" "error" "$@"
}

cc::console::warning(){
  _cc_private::console "$CC_COLOR_YELLOW" "$CC_COLOR_BLACK" "warning" "$@"
}

cc::console::info(){
  _cc_private::console "$CC_COLOR_GREEN" "$CC_COLOR_BLACK" "info" "$@"
}

cc::console::debug(){
  _cc_private::console "$CC_COLOR_WHITE" "$CC_COLOR_BLACK" "debug" "$@"
}

cc::console::body(){
  _cc_private::console "$CC_COLOR_WHITE" "" "" "$@"
}

cc::console::comment(){
  _cc_private::console "$CC_COLOR_CYAN" "" "" "$@"
}

cc::console::humpf(){
  _cc_private::console "$CC_COLOR_MAGENTA" "" "" "$@"
}

cc::console::reset(){
  _cc_private::tput setaf $CC_COLOR_DEFAULT_FRONT
  _cc_private::tput setab $CC_COLOR_DEFAULT_BACK
  printf "\n"
}

cc::console::end(){
  _cc_private::tput op
  printf "\n"
}

_CC_PRIVATE_LOGGER_LEVEL=2

_cc_private::logger::log(){
  local prefix="$1"
  shift

  local level="CC_LOGGER_$prefix"
  local i

  [ "$_CC_PRIVATE_LOGGER_LEVEL" -ge "${!level}" ] || return 0

  # About the crazy shit: https://stackoverflow.com/questions/12674783/bash-double-process-substitution-gives-bad-file-descriptor
  exec 3>&2
  for i in "$@"; do
    >&2 cc::console::"$(printf "$prefix" | tr '[:upper:]' '[:lower:]')" printf "$i"
  done
  exec 3>&-
}

cc::logger::level::set() {
  local level
  level="$(printf "%s" "${1:-}" | tr '[:upper:]' '[:lower:]')"

  case "$level" in
    ''|*[!0-9]*)
      case "$level" in
          "debug")
            _CC_PRIVATE_LOGGER_LEVEL=4
            ;;
          "info")
            _CC_PRIVATE_LOGGER_LEVEL=3
            ;;
          "warning")
            _CC_PRIVATE_LOGGER_LEVEL=2
            ;;
          "error")
            _CC_PRIVATE_LOGGER_LEVEL=1
            ;;
          "mute")
            _CC_PRIVATE_LOGGER_LEVEL=0
            ;;
          *)
            _CC_PRIVATE_LOGGER_LEVEL=3
            ;;
      esac
      ;;
    *)
      _CC_PRIVATE_LOGGER_LEVEL="$level"
      ;;
  esac

  [ "$_CC_PRIVATE_LOGGER_LEVEL" != "$CC_LOGGER_DEBUG" ] || {
    cc::console::warning printf "YOU ARE LOGGING AT THE DEBUG LEVEL."
    cc::console::warning printf "This is NOT recommended for production use, and WILL LIKELY LEAK sensitive information to logs."
  }
}

# Sugar
cc::logger::level::set::debug(){
  cc::logger::level::set "$CC_LOGGER_DEBUG"
}

cc::logger::level::set::info(){
  cc::logger::level::set "$CC_LOGGER_INFO"
}

cc::logger::level::set::warning(){
  cc::logger::level::set "$CC_LOGGER_WARNING"
}

cc::logger::level::set::error(){
  cc::logger::level::set "$CC_LOGGER_ERROR"
}

cc::logger::mute() {
  # shellcheck disable=SC2034
  _CC_PRIVATE_LOGGER_LEVEL=0
}

cc::logger::ismute() {
  # shellcheck disable=SC2034
  [ "$_CC_PRIVATE_LOGGER_LEVEL" == 0 ] || return "$ERROR_GENERIC_FAILURE"
}

cc::logger::debug(){
  _cc_private::logger::log "DEBUG" "$@"
}

cc::logger::info(){
  _cc_private::logger::log "INFO" "$@"
}

cc::logger::warning(){
  _cc_private::logger::log "WARNING" "$@"
}

cc::logger::error(){
  _cc_private::logger::log "ERROR" "$@"
}

cc::tracer(){
  local length="${#BASH_SOURCE[@]}"

  #local indent=""
  local linenumber
  local filename
  local filecontent
  local funcname
  local col

  [ "$CC_DEBUG_CORE" == "true" ] || [ "${BASH_SOURCE[1]}" != "${BASH_SOURCE[0]}" ] || return 0
  >&2 printf "\n"

  for (( j=$(( length - 1 )); j>0; j-- )); do
    [ "$CC_DEBUG_CORE" == "true" ] || [ "${BASH_SOURCE[$j]}" != "${BASH_SOURCE[0]}" ] || {
      continue
    }

    linenumber="${BASH_LINENO[$(( j - 1 ))]}"
    filename="${BASH_SOURCE[$j]}"
    filecontent="$(cat -n "$filename" | grep -E "^\s+$linenumber\s" | sed -E "s/^\s+$linenumber\s+//")"
    funcname="${FUNCNAME[$j]}"

    # Why is this showing is a mystery for the times
    # [ "$filecontent" != "#!/usr/bin/env bash" ] || continue

    [ "$j" != 1 ] && {
      col=$CC_COLOR_WHITE
    } || {
      col=$CC_COLOR_GREEN
    }
    >&2 _cc_private::console::inline "$col" "$CC_COLOR_BLACK" "file   " printf "%-35s" "$filename"
    >&2 _cc_private::console::inline "$col" "$CC_COLOR_BLACK" "line" printf "%-9s" "#$linenumber"
    >&2 _cc_private::console::inline "$col" "$CC_COLOR_BLACK" "function" printf "%-30s" "$funcname()"
    >&2 printf "\n"
    [ "$j" != 1 ] || {
      >&2 _cc_private::console "$CC_COLOR_BLUE" "$CC_COLOR_WHITE" "command" printf "%s" "$filecontent"
      >&2 _cc_private::console "$CC_COLOR_RED" "$CC_COLOR_WHITE" "output" printf ""
    }

  done
}

_CC_PRIVATE_TRAP_LISTENERS=()

_CC_PRIVATE_ERR_LNO=
_CC_PRIVATE_ERR_CMD=
_CC_PRIVATE_ERR_STACK=
_CC_PRIVATE_ERR_EX=

dc::trap::register(){
  _CC_PRIVATE_TRAP_LISTENERS+=( "$1" )
}

# Trap lno and cmd on ERR for future use
cc::trap::err(){
  _CC_PRIVATE_ERR_EX="$1"
  _CC_PRIVATE_ERR_LNO="$2"
  _CC_PRIVATE_ERR_CMD="$3"
  _CC_PRIVATE_ERR_STACK="$4"
  # Dropping the rest of the stack?

  >&2 printf "\n"
  >&2 _cc_private::console::inline "$CC_COLOR_RED" "$CC_COLOR_BLACK" "file   " printf "%-35s" "$_CC_PRIVATE_ERR_STACK"
  >&2 _cc_private::console::inline "$CC_COLOR_RED" "$CC_COLOR_BLACK" "line" printf "%-9s" "#$_CC_PRIVATE_ERR_LNO"
  >&2 _cc_private::console::inline "$CC_COLOR_RED" "$CC_COLOR_BLACK" "command" printf "%-30s" "$_CC_PRIVATE_ERR_CMD"
  >&2 _cc_private::console::inline "$CC_COLOR_RED" "$CC_COLOR_BLACK" "exit" printf "%s" "$_CC_PRIVATE_ERR_EX"
  >&2 _cc_private::tput setaf "$CC_COLOR_DEFAULT_FRONT"
  >&2 printf "\n"

  >&2 _cc_private::console "$CC_COLOR_RED" "$CC_COLOR_WHITE" "source" printf ""
  prefix=" "
  for (( j=$(( _CC_PRIVATE_ERR_LNO - 3 )); j<$(( _CC_PRIVATE_ERR_LNO + 3 )); j++ )); do
    [ "$j" -gt 0 ] || continue
    if [ "$j" == "$_CC_PRIVATE_ERR_LNO" ]; then
      prefix=">"
      >&2 _cc_private::tput setaf "$CC_COLOR_RED"
      #>&2 _cc_private::tput bold
    fi
    #  | sed -E "s/^\s+$j\s//"
    >&2 printf "%s%s" "$prefix" "$(cat -n "$_CC_PRIVATE_ERR_STACK" | grep -E "^\s+$j\s")" || true
    if [ "$j" == "$_CC_PRIVATE_ERR_LNO" ]; then
      prefix=" "
      >&2 _cc_private::tput setaf "$CC_COLOR_DEFAULT_FRONT"
      #>&2 _cc_private::tput sgr0
    fi
    >&2 printf "\n"
  done
}

_DC_NO_REENTRY=

# Trap exit for the actual cleanup
cc::trap::exit() {
  { set +x; } 2>/dev/null

  local ex="$1"
  local i

  # Prevent reentrancy - XXX is this actually needed?
  [ ! "$_DC_NO_REENTRY" ] || return 0
  _DC_NO_REENTRY="exiting"

  printf "%s\n" "$ex" > "$CC_TMPFS"/.codecomet/logs/ex.log

  if [ "${#_CC_PRIVATE_TRAP_LISTENERS[@]}" -gt 0 ]; then
    for i in "${_CC_PRIVATE_TRAP_LISTENERS[@]}"; do
        cc::logger::debug "Calling exit hook $i"
#      >&2 printf "\n"
      "$i" "$ex" "$_CC_PRIVATE_ERR_CMD" "$_CC_PRIVATE_ERR_LNO" "$_CC_PRIVATE_ERR_STACK"
    done
  fi
  >&2 cc::console::reset
  >&2 printf "\n"
  cc::logger::debug "Exiting ($ex)"
  exit "$ex"
}


