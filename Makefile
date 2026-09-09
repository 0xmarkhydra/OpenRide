.PHONY: stack-up stack-down stack-logs stack-ps stack-reset infra-up infra-down api-run api-test api-fmt core-test core-fmt core-example modules-test modules-fmt contracts-test sdk-test packages-test workspace-test docker-test docker-integration-test docker-test-all

stack-up:
	docker compose up -d --build --wait

stack-down:
	docker compose down --remove-orphans

stack-logs:
	docker compose logs -f --tail=200

stack-ps:
	docker compose ps

# Destructive, but scoped only to the OpenRide Compose project.
stack-reset:
	docker compose down -v --remove-orphans

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

workspace-test: packages-test api-test

# Ephemeral containers: --rm guarantees no test container is left behind.
docker-test:
	docker compose --profile test run --rm --no-deps --build api-test

docker-integration-test:
	docker compose up -d --wait postgres redis
	docker compose --profile test run --rm --no-deps --build api-integration

docker-test-all: docker-test docker-integration-test
