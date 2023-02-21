#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail

go run codecomet/builders/debian/debian.go "$@"
./send.sh "cc-builders-debian" "Codecomet: building debian base images"
