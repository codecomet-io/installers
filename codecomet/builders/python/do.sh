#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail

go run codecomet/builders/python/python.go "$@"
./send.sh "cc-builders-python" "Building python base images"
