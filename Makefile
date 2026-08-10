.PHONY: infra-up infra-down api-run api-test api-fmt

infra-up:
	docker compose up -d

infra-down:
	docker compose down

api-run:
	cd services/api && go run ./cmd/api

api-test:
	cd services/api && go test ./...

api-fmt:
	cd services/api && gofmt -w .
