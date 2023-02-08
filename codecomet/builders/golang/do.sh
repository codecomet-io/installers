#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail

go run codecomet/builders/golang/golang.go build
./send.sh "cc-builders-golang" "Building golang base images"
