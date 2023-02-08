#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail

go run codecomet/builders/c/c.go build "$@"
./send.sh "cc-builders-c" "Codecomet: building c base image"
