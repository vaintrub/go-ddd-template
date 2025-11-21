#!/bin/bash
set -e

if [ "$2" == "-install" ]; then
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
  go install github.com/roblaszczak/go-cleanarch@latest
fi

readonly service="$1"

cd "./internal/$service"
golangci-lint run