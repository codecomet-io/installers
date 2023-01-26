#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail -o monitor

cc::init::speedup(){
  ! command -v eatmydata >/dev/null || export LD_PRELOAD=libeatmydata.so
}

cc::bootstrap(){
  local scriptFilename="$1"
  local base
  base="$(dirname "${BASH_SOURCE[0]}")"

  local out="$CC_TMPFS"/.codecomet/logs/stdout.log
  local err="$CC_TMPFS"/.codecomet/logs/stderr.log

  cp "$base/$scriptFilename" "$CC_TMPFS/.codecomet/bin"
  scriptFilename="$CC_TMPFS/.codecomet/bin/$scriptFilename"

  # Generate starter script as well for future use
  # echo "$scriptFilename" > "$CC_TMPFS"/.codecomet/bin/cc-redo

  # shellcheck disable=SC1090
  source "$scriptFilename" > >(tee -a "$out") 2> >(tee -a "$err" >&2)
}

ccdebug_stdout(){
  if [ "${1:-}" == "-f" ]; then
    tail -f "$CC_TMPFS"/.codecomet/logs/stdout.log
  else
    cat "$CC_TMPFS"/.codecomet/logs/stdout.log
  fi
}

ccdebug_stderr(){
  if [ "${1:-}" == "-f" ]; then
    tail -f "$CC_TMPFS"/.codecomet/logs/stderr.log
  else
    cat "$CC_TMPFS"/.codecomet/logs/stderr.log
  fi
}

ccdebug_ex(){
  [ -e "$CC_TMPFS"/.codecomet/logs/ex.log ] && {
    cat "$CC_TMPFS"/.codecomet/logs/ex.log
  } || {
    cc::logger::warning "Action is in progress, no exit code yet"
  }
}

ccdebug_action(){
  echo "$_CC_PRIVATE_SCRIPT"
}

cc::init::tracer(){
  # shellcheck disable=SC2016
  local ps=('$(' "$@" ')')
  export PS4="${ps[*]}"
}

# Forking to disable xtrace when entering traps
cc::init::trap(){
  # Signals are caught by buildkit already - and only SIGKILL triggers a failure
  trap '{ ex=$?; set +x; } 2>/dev/null && cc::trap::err "$ex" "$LINENO" "$BASH_COMMAND" "${BASH_SOURCE[@]}"' ERR
  trap '{ ex=$?; set +x; } 2>/dev/null && cc::trap::exit "$ex"' EXIT
}

# Boot if we have an argument - otherwise, we are being sourced
if [ "$#" -gt 0 ]; then
  rm -Rf "$CC_TMPFS"/.codecomet
  mkdir -p "$CC_TMPFS"/.codecomet/bin
  mkdir -p "$CC_TMPFS"/.codecomet/logs
  mkdir -p "$TMPDIR"

  export _CC_PRIVATE_SCRIPT="$CC_TMPFS/.codecomet/bin/$1"

  # Set logger to env var from the Bash helper
  cc::logger::level::set "$CC_DEBUG_LEVEL"

  # Eat data, trap, register tracer, register debugger on exit trap
  cc::init::speedup
  cc::init::trap
  cc::init::tracer cc::tracer
  if [ "${CC_DEBUG_LIVE:-}" != "" ]; then
    cc::debugger::live &
    dc::trap::register cc::debugger::front
  else
    dc::trap::register cc::debugger::start
  fi

  # Boot it
  #_CC_PRIVATE_SCRIPT=
  #export _CC_PRIVATE_SCRIPT
  cc::bootstrap "$@"
else
  # If we are a library, toss this one so we do not exit on any error...
  set +o errexit
fi

