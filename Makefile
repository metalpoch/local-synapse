include .env

.PHONY: help build-mcp build-api run-api test docker-build docker-run

help:
	@echo "Personal MCP Tools Repository - Makefile"
	@echo ""
	@echo "Commands:"
	@echo "  make help           - Show this help message"
	@echo "  make build-mcp      - Build MCP server binary"
	@echo "  make build-api      - Build API server binary"
	@echo "  make run-api        - Run API server"
	@echo "  make test           - Run tests"

build-mcp:
	go build -o mcp ./cmd/mcp/main.go

build-api:
	go build -o synapse ./cmd/api/main.go

test:
	go test ./...

run-api: build-mcp build-api
	PORT=$(PORT) \
	OLLAMA_URL=${OLLAMA_URL} \
	OLLAMA_MODEL=$(OLLAMA_MODEL) \
	OLLAMA_SYSTEM_PROMPT=$(OLLAMA_SYSTEM_PROMPT) \
	./synapse
