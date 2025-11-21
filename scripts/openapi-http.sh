#!/bin/bash
set -euo pipefail

if ! command -v oapi-codegen >/dev/null 2>&1; then
	echo "oapi-codegen binary not found; install via 'go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest'" >&2
	exit 1
fi

readonly service="$1"
readonly output_dir="$2"
readonly package="$3"

oapi-codegen -generate types -o "$output_dir/openapi_types.gen.go" -package "$package" "api/openapi/$service.yml"
oapi-codegen -generate chi-server -o "$output_dir/openapi_api.gen.go" -package "$package" "api/openapi/$service.yml"
oapi-codegen -generate types -o "internal/common/client/$service/openapi_types.gen.go" -package "$service" "api/openapi/$service.yml"
oapi-codegen -generate client -o "internal/common/client/$service/openapi_client_gen.go" -package "$service" "api/openapi/$service.yml"
