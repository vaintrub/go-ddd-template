include .env
export

.PHONY: openapi_http
openapi_http:
	@./scripts/openapi-http.sh trainer internal/trainer/ports ports
	@./scripts/openapi-http.sh trainings internal/trainings/ports ports
	@./scripts/openapi-http.sh users internal/users main

.PHONY: proto
proto:
	@./scripts/proto.sh trainer
	@./scripts/proto.sh users

.PHONY: lint
lint:
	@go-cleanarch -interfaces ports -infrastructure adapters
	@./scripts/lint.sh common
	@./scripts/lint.sh trainer
	@./scripts/lint.sh trainings
	@./scripts/lint.sh users

.PHONY: fmt
fmt:
	goimports -l -w internal/

.PHONY: c4
c4:
	cd tools/c4 && go mod tidy && sh generate.sh

test:
	@./scripts/test.sh common .e2e.env
	@./scripts/test.sh trainer .test.env
	@./scripts/test.sh trainings .test.env
	@./scripts/test.sh users .test.env
