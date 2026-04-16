.PHONY: help tidy test test-unit test-integration mocks tools docker-up docker-down docker-logs

help:
	@echo "Targets:"
	@echo "  tidy              - go mod tidy"
	@echo "  test              - run all tests"
	@echo "  test-unit         - run unit tests (fast, no docker)"
	@echo "  test-integration  - run integration tests (requires docker)"
	@echo "  tools             - install dev tools (mockery)"
	@echo "  mocks             - (re)generate mocks with mockery"
	@echo "  docker-up         - docker compose up --build -d"
	@echo "  docker-down       - docker compose down"
	@echo "  docker-logs       - docker compose logs -f --tail=200"

tidy:
	go mod tidy

test:
	go test ./... -count=1 -v

test-unit:
	go test ./internal/application/... ./internal/adapters/http/... -count=1 -v

test-integration:
	go test ./tests/integration/... -count=1 -v

tools:
	go install github.com/vektra/mockery/v3@latest

mocks:
	@echo "Generating mocks..."
	mockery --config .mockery.yml

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f --tail=200

