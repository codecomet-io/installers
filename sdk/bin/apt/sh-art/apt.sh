#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail -o monitor

# Helpers to allow apt to share cache properly

# Private tmpfs location
_cc_tmpfs_location=/codecomet/apt-get-private/tmp

# Location of the list cache
_cc_apt_list_store_location=/codecomet/apt-get-shared/lists
# And pack cache
_cc_apt_pack_local_location=/codecomet/apt-get-shared/packs

# Location of the packages cache
_cc_apt_pack_store_location=/codecomet/apt-get-shared/cache

# Location of our private lists files - by default, Debian normal location, which means state will persist
_cc_apt_list_local_location=/var/lib/apt/lists

# Location of the config file
export APT_CONFIG=$_cc_tmpfs_location/apt-get.conf

# Hash of the sources.list, which dictates the cache location to share
# Different sources means different cache location
# Caveats: this will likely not work with a future debian release as they change the sources.list format
# XXX this wracking the debugger output
cc::apt_get::shard(){
  sha256sum <<< "$(uname -m)$(grep --no-filename -Ev "^#" /etc/apt/sources.list /etc/apt/sources.list.d/* 2>/dev/null | sed -E 's/#.+$//' | sort || true)" | sed -E 's/  .*//'
}

# Set the tmp location, initiliaze APT_CONFIG and shared and private locations
cc::apt_get::init(){
  # Get the desired tmpfs mount point
  _cc_tmpfs_location="$1"

  mkdir -p "$_cc_tmpfs_location"

  # Point configuration there
  export APT_CONFIG="$_cc_tmpfs_location"/apt-get.conf

  # Sharding by sources list content, cleaned-up and sorted to maximize cache hits
	_cc_apt_list_store_location="$2/$(cc::apt_get::shard)"

  # Technically, we should not be sharding the architecture
  # But then, the only upside would be when installing cross-arch packages
	_cc_apt_pack_store_location="$3/$(cc::apt_get::shard)"

  # If we do not want lists to persist, use the tmp storage
  [ "${4:-}" == true ] || _cc_apt_list_local_location="$_cc_tmpfs_location"/lists

  # Finally, local pack location
  _cc_apt_pack_local_location="$_cc_tmpfs_location"/packs

  # Make sure they all exist
  mkdir -p "$_cc_apt_list_local_location"
  mkdir -p "$_cc_apt_pack_local_location"

  mkdir -p "$_cc_apt_list_store_location"
  mkdir -p "$_cc_apt_pack_store_location"
}

# Create the configuration file for APT
cc::apt_get::configure(){
  local config_extra="${1:-}"
  local persist="${2:-}"

  mkdir -p "$_cc_tmpfs_location/logs"
  cat << EOF > "$APT_CONFIG"
# Divert both lists and archives location
Dir::Cache::Archives "$_cc_apt_pack_local_location";
Dir::State::Lists "$_cc_apt_list_local_location";
# Prevent the default apt config to be used
Dir::Etc::parts "";
# Put the logs in the tmp zone
Dir::Log "$_cc_tmpfs_location/logs";
# Put the bin cache files in the private mount
Dir::Cache::srcpkgcache "$_cc_tmpfs_location/srcpkgcache.bin";
Dir::Cache::pkgcache "$_cc_tmpfs_location/pkgcache.bin";

$config_extra
EOF

  # If asked to persist the config, copy over the generated file
  [ "$persist" == "false" ] || cp "$APT_CONFIG" /etc/apt/apt.conf
}

#cc::apt_get::cleanup(){
  # This is only necessary if we were not using an tmpfs mount...
  # rm "$APT_CONFIG"
#}

cc::apt_get::update(){
  # Remove our marker if here
  rm -f "$_cc_apt_list_local_location"/.cc_updated

  # Retrieve existing data
  cc::storage::retrieve "$_cc_apt_list_store_location" "$_cc_apt_list_local_location" copy

	# Do the deed, same restriction applies
	apt-get update "$@"

  # Save it back (storage will decide)
  cc::storage::store "$_cc_apt_list_local_location" "$_cc_apt_list_store_location" erase

  # Now, flag it
  date > "$_cc_apt_list_local_location"/.cc_updated
}

cc::apt_get::do(){
  # Bring in any cache
  cc::storage::retrieve "$_cc_apt_pack_store_location" "$_cc_apt_pack_local_location"

	apt-get "$@"

  # Store the packages permanently
  cc::storage::store "$_cc_apt_pack_local_location" "$_cc_apt_pack_store_location"
}

# XXX what if sources are being modified, or config, or something else that materially impacts update?
cc::high::update(){
  # [ -e "$_cc_apt_list_local_location"/.cc_updated ] || time cc::apt_get::update -qq
  time cc::apt_get::update -qq
}

cc::high::install(){
  # Do have a marker indicating that the lists has been kept around?
  # If not, then force update
  [ -e "$_cc_apt_list_local_location"/.cc_updated ] || time cc::apt_get::update -qq

  time cc::apt_get::do install -qq "$@"
}

cc::high::upgrade(){
  # Do have a marker indicating that the lists have been kept around?
  # If not, then force update
  [ -e "$_cc_apt_list_local_location"/.cc_updated ] || time cc::apt_get::update -qq

  # ls -lAR "$_cc_tmpfs_location"/packs || true
  # time cc::apt_get::do upgrade -qq "$@"

  # XXX WIP here
  # Bust cache - delete this
  #time cc::apt_get::do upgrade --download-only -qq "$@"
  #time cc::apt_get::do upgrade --no-download -qq "$@"
  time cc::apt_get::do upgrade -qq "$@"
}

cc::high::purge(){
  time cc::apt_get::do purge -qq --auto-remove "$@"
}