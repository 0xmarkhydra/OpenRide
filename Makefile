.PHONY: stack-up stack-down stack-logs stack-ps stack-reset micro-up micro-down micro-logs micro-config infra-up infra-down api-run api-test api-fmt marketplace-run marketplace-test marketplace-fmt core-test core-fmt core-example modules-test modules-fmt contracts-test sdk-test packages-test services-test workspace-test docker-test docker-integration-test docker-test-all

stack-up:
	docker compose up -d --build --wait

stack-down:
	docker compose down --remove-orphans

stack-logs:
	docker compose logs -f --tail=200

stack-ps:
	docker compose ps

# Destructive, but scoped only to the legacy compatibility Compose project.
stack-reset:
	docker compose down -v --remove-orphans

micro-up:
	docker compose -f docker-compose.microservices.yml up -d --build --wait

micro-down:
	docker compose -f docker-compose.microservices.yml down --remove-orphans

micro-logs:
	docker compose -f docker-compose.microservices.yml logs -f --tail=200

micro-config:
	docker compose -f docker-compose.microservices.yml config --quiet

# Backward-compatible infrastructure aliases.
infra-up:
	docker compose up -d --wait postgres redis

infra-down:
	docker compose stop postgres redis

api-run:
	cd services/api && go run ./cmd/api

api-test:
	cd services/api && go test ./...

api-fmt:
	cd services/api && gofmt -w .

marketplace-run:
	cd services/marketplace && go run ./cmd/marketplace

marketplace-test:
	cd services/marketplace && go test ./...

marketplace-fmt:
	cd services/marketplace && gofmt -w .

core-test:
	cd packages/core-go && go test ./...

core-fmt:
	cd packages/core-go && gofmt -w .

core-example:
	cd packages/core-go && go run ./examples/minimal

modules-test:
	cd packages/modules-go && go test ./...

modules-fmt:
	cd packages/modules-go && gofmt -w .

contracts-test:
	cd packages/contracts && npm run validate && npm pack --dry-run >/dev/null

sdk-test:
	cd packages/sdk && npm test && npm pack --dry-run >/dev/null

packages-test: core-test modules-test contracts-test sdk-test

services-test: marketplace-test api-test

workspace-test: packages-test services-test

# Ephemeral containers: --rm guarantees no test container is left behind.
docker-test:
	docker compose --profile test run --rm --no-deps --build api-test

docker-integration-test:
	docker compose up -d --wait postgres redis
	docker compose --profile test run --rm --no-deps --build api-integration

docker-test-all: docker-test docker-integration-test
