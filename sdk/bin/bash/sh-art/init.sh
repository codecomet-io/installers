#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail -o monitor

# See https://manpages.ubuntu.com/manpages/xenial/man1/eatmydata.1.html
# Technically, we get a 10% speedup with apt operations for eg
# XXX careful here if this is going to be used in a chroot
cc::init::speedup(){
  ! command -v eatmydata >/dev/null || export LD_PRELOAD=libeatmydata.so
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

ccdebug_stdex(){
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
  mkdir -p "$TMPDIR"
  rm -Rf "$CC_TMPFS"/.codecomet
  mkdir -p "$CC_TMPFS"/.codecomet/bin
  mkdir -p "$CC_TMPFS"/.codecomet/logs

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

  # XXX technically, we receive a bunch of scripts, and we could just play them all - use case is not completely clear yet
  out="$CC_TMPFS"/.codecomet/logs/stdout.log
  err="$CC_TMPFS"/.codecomet/logs/stderr.log

  while [ "$#" -gt 1 ]; do
    cp "$1" "$CC_TMPFS"/.codecomet/bin
    # shellcheck disable=SC1090
    source "$CC_TMPFS"/.codecomet/bin/"$(basename "$1")"
    shift
  done
  cp "$1" "$CC_TMPFS"/.codecomet/bin
  export _CC_PRIVATE_SCRIPT="$CC_TMPFS"/.codecomet/bin/"$(basename "$1")"

  # Boot it
  # shellcheck disable=SC1090
  source "$_CC_PRIVATE_SCRIPT" > >(tee -a "$out") 2> >(tee -a "$err" >&2)
else
  # If we are a library, toss this one so we do not exit on any error...
  # This is especially important for the reverse debugger
  set +o errexit
fi

