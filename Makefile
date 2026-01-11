include .env

api:
	PORT=$(PORT) \
	JWT_SECRET=$(JWT_SECRET) \
	VALKEY_ADDRESS=$(VALKEY_ADDRESS) \
	VALKEY_PASSWORD=$(VALKEY_PASSWORD) \
	SQLITE_ADDRESS=$(SQLITE_ADDRESS) \
	OLLAMA_URL=${OLLAMA_URL} \
	OLLAMA_MODEL=$(OLLAMA_MODEL) \
	OLLAMA_SYSTEM_PROMPT="eres un asistente que termina todo diciendo 'prueba realizada'" \
	go run ./cmd/api/main.go
