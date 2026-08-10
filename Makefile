.PHONY: stack-up stack-down stack-logs stack-ps stack-reset infra-up infra-down api-run api-test api-fmt docker-test docker-integration-test docker-test-all

stack-up:
	docker compose up -d --build --wait

stack-down:
	docker compose down --remove-orphans

stack-logs:
	docker compose logs -f --tail=200

stack-ps:
	docker compose ps

# Destructive, but scoped only to the FlashX Compose project.
stack-reset:
	docker compose down -v --remove-orphans

# Backward-compatible aliases.
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

# Ephemeral containers: --rm guarantees no test container is left behind.
docker-test:
	docker compose --profile test run --rm --no-deps --build api-test

docker-integration-test:
	docker compose up -d --wait postgres redis
	docker compose --profile test run --rm --no-deps --build api-integration

docker-test-all: docker-test docker-integration-test
