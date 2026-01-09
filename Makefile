include .env

run-api:
	JWT_SECRET=$(JWT_SECRET) \
	OLLAMA_MODEL=$(OLLAMA_MODEL) \
	PORT=$(PORT) \
	go run ./cmd/api/main.go
