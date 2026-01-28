package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
	mcpclient "github.com/metalpoch/local-synapse/internal/infrastructure/mcp_client"
	"github.com/metalpoch/local-synapse/internal/usecase/ollama"
)

func SetupOllamaRouter(e *echo.Echo, ollamaUrl, model, systemPrompt string, mcpClient mcpclient.MCPClient) {
	h := handler.NewOllamaHandler(
		ollama.NewStreamChatUsecase(ollamaUrl, model, systemPrompt, mcpClient),
	)

	router := e.Group("/api/v1/ollama")
	router.GET("/chat", h.Stream)
}
