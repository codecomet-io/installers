#!/usr/bin/env bash
set -o errexit -o errtrace -o functrace -o nounset -o pipefail

go run codecomet/builders/node/node.go build
./send.sh "cc-builders-node" "Building node base images"
