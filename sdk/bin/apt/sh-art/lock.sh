#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail -o monitor

# A generic lock mechanism that supports exclusive locks, shared locks, and provide a queue mechanism
_cc_lockindex=9
_cc_lock_location="/_cc/share/locks"

# Sets our locks location. Must be called to guarantee that the lock location exist.
cc::lock::init(){
	_cc_lock_location="${1:-.}"
	mkdir -p "${_cc_lock_location}"
}

# Acquire a lock with a given name, and a share mode (shared or exclusive (which is the default))
# If the lock is already acquired in a different mode, or the mode is exclusive, the function will wait until the lock is released.
# Return the lock id that you need to carry around to release the lock later on.
cc::lock::acquire(){
	local lockfile="$_cc_lock_location/$1"
	local shared="${2:-}"
	[ "$shared" ] && shared=-s || shared=-x
	_cc_lockindex=$((_cc_lockindex+1))
	mkdir -p "$(dirname "$lockfile")"
	exec {_cc_lockindex}>"$lockfile"
	flock $shared $_cc_lockindex
	return $_cc_lockindex
}

# Release a previously acquire lock by its id
cc::lock::release(){
	local idx="$1"
	# This is essentially the same
	flock -u "$idx"
	# exec {idx}>&-
}

cc::lock::queue(){
	local basepath="${1:-.}"
	local shared1="${2:-}"
	local shared2="${3:-}"
	local qid
	local oid

	# Get into the queue first
	cc::lock::acquire "$basepath"/cc_queue.lock "$shared1" || qid=$?

	# Then when out of the queue, acquire an operation lock
	cc::lock::acquire "$basepath"/cc_op.lock "$shared2" || oid=$?

	# Now, leave the queue
	cc::lock::release $qid

	# Return the lock id
	return $oid
}
