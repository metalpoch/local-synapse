package router

import (
	"context"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/infrastructure/mcp_client"
	"github.com/valkey-io/valkey-go"
)

type Config struct {
	Echo               *echo.Echo
	Cache              *valkey.Client
	Secret             string
	OllamaUrl          string
	OllamaModel        string
	OllamaSystemPrompt string
}

func Init(cfg *Config) {
	// Initialize MCP Client (local binary)
	mcpClient, err := mcpclient.NewStdioClient("./mcp")
	if err != nil {
		log.Printf("Failed to create MCP client: %v", err)
	} else {
		if err := mcpClient.Initialize(context.Background()); err != nil {
			log.Printf("Failed to initialize MCP client: %v", err)
		}
	}

	SetupSystemRouter(cfg.Echo)
	SetupOllamaRouter(cfg.Echo, cfg.OllamaUrl, cfg.OllamaModel, cfg.OllamaSystemPrompt, mcpClient)
}
