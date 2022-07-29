#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail

readonly ISOV_PATH="$HOME/Projects/GitHub/codecomet/isovaline"
readonly INST_PATH="$(cd "$(dirname "${BASH_SOURCE[0]:-$PWD}")" 2>/dev/null 1>&2 && pwd)"/..
readonly CC_VERSION="mark-I"

build(){
  rm -Rf dist
  make all
  rm dist/share/lima/lima-guestagent.Linux-riscv64

  local destination="$INST_PATH/release/$CC_VERSION/$GOOS/$GOARCH"
  rm -Rf "$destination"
  mkdir -p "$(dirname "$destination")"
  cp -R dist "$destination"
}

cd "$ISOV_PATH" || exit

export GOOS=darwin
export GOARCH=arm64
build

export GOOS=darwin
export GOARCH=amd64
build

cd - > /dev/null || exit