#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail -o monitor

# Lessons learned:
# - find -exec is VERY SLOW - better off piping to read (almost 10x) (0.2 second vs 0.02 second for about 200 files)
# - then globbing and grepping is faster than find + read
# This matters not in case we are copying a large amount of data of course, but then
# - there is no way to get something stable out of ls without further processing - directories with different owners
# for eg will widen some columns
# this "works" right now (if we do not use --numeric-uid-gid) but will be problematic outside of this specific case

_cc::fingerprint(){
  local directory="$1"
  # Will fail if the directory is empty, so the guardrail
  # Ignore directories, links, and "total" (note: -d will not work for that)
  # --numeric-uid-gid < messes up the output width
  # shellcheck disable=SC2010
  ls --full-time --almost-all --ignore lock "$directory" 2>/dev/null | grep -v "^[d|l|t]" || true
}

# Copy or symlink data "from" storage "to" destination
# This assume that storage holds only files, and no lock
# Will not copy anything if there is no difference
cc::storage::retrieve(){
  local from="$1"
  local to="$2"
  local copy="${3:-}"
  local lid
  local fromState

  # Ensure destination exists
  mkdir -p "$to"

  # Lock origin and keep the lock id
  cc::lock::acquire "$from" shared || lid=$?

  # Ensure origin exists
  mkdir -p "$from"

  # If we are asked to copy (not expected for large amount of data - linking is prefered in that case)
  if [ "$copy" ]; then
    # Retrieve the state of it
    fromState="$(_cc::fingerprint "$from")"

    # Check that we have content in storage
    # Check that we have a difference between storage and destination (ignoring directories and lock file)
    # shellcheck disable=SC2010
    # shellcheck disable=SC2015
    [ "$fromState" ] && [ "$fromState" != "$(_cc::fingerprint "$to")" ] && {
      # Then cleanup destination
      rm -Rf "$to"
      mkdir -p "$to"
      # And copy over
      cp -p "$from"/* "$to"
      # Slower
      #find "$from" -type f -not -iname lock -print0 | while IFS= read -r -d $'\0' fd; do
      #  cp -p "$fd" "$to"
      #done
    } || {
      # Otherwise, do nothing
      cc::logger::debug "Nothing to retrieve from storage"
    }
  else
    # We want links, so, lets check we have anything in the origin
    #[ ! "$fromState" ] || {
    # Get rid of possibly remaining files in our destination
    find "$to" -type f -delete
    ln -sf "$from"/* "$to" 2>/dev/null || true
    # Note: below is slower
    # find "$from" -type f -not -iname lock -print0 | while IFS= read -r -d $'\0' fd; do
    #  ln -s "$fd" "$to"
    # done
    #}
  fi

  # Release the lock
	cc::lock::release $lid
}

cc::storage::store(){
  local from="$1"
  local to="$2"
  local erase="${3:-}"
  local lid
  local fd
  local toState

  # Ensure origin exists
  mkdir -p "$from"

  # Lock destination and keep the lock id
  cc::lock::acquire "$to" "" || lid=$?

  # Ensure destination exists
  mkdir -p "$to"
  # Retrieve the state of it
  # shellcheck disable=SC2010
  toState="$(_cc::fingerprint "$to")"

  # Compare state of the destination with the origin (origin ignores directories and lock files)
  # shellcheck disable=SC2010
  # shellcheck disable=SC2015
  [ "$toState" != "$(_cc::fingerprint "$from")" ] && {
    # We have changes - if asked to wipe out, do so
    if [ "$erase" ]; then
      rm -Rf "$to"
      mkdir -p "$to"
    fi
    # Now copy the files
    # This is probably slow-ish. The question is: will it be significant in a context where copy IO is the bottleneck?
    find "$from" -type f -not -iname "lock" -print0 | while IFS= read -r -d $'\0' fd; do
      cp -p "$fd" "$to"
    done
  } || {
    cc::logger::debug "Nothing to save to storage"
  }

  # Release the lock
	cc::lock::release $lid
}
