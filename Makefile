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

# Database Migration Targets
.PHONY: migrate-up
migrate-up:
	@migrate -path migrations -database "$(DATABASE_URL)" -verbose up

.PHONY: migrate-down
migrate-down:
	@migrate -path migrations -database "$(DATABASE_URL)" -verbose down 1

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME parameter is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir migrations -seq $(NAME)

.PHONY: migrate-status
migrate-status:
	@migrate -path migrations -database "$(DATABASE_URL)" version

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(V)" ]; then \
		echo "Error: V parameter is required. Usage: make migrate-force V=version_number"; \
		exit 1; \
	fi
	@migrate -path migrations -database "$(DATABASE_URL)" force $(V)

# SQLC Code Generation
.PHONY: sqlc-generate
sqlc-generate:
	@sqlc generate

# Casdoor Docker Compose helpers
CASDOOR_COMPOSE_FILE ?= docker/casdoor/docker-compose.casdoor.yml

.PHONY: casdoor-up
casdoor-up:
	@docker compose -f $(CASDOOR_COMPOSE_FILE) --env-file .cusdoor.env up -d

.PHONY: casdoor-down
casdoor-down:
	@docker compose -f $(CASDOOR_COMPOSE_FILE) --env-file .cusdoor.env down

.PHONY: casdoor-logs
casdoor-logs:
	@docker compose -f $(CASDOOR_COMPOSE_FILE) --env-file .cusdoor.env logs -f
