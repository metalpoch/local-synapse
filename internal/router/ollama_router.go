package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/handler"
)

func SetupOllamaRouter(e *echo.Echo, ollamaUrl, model, systemPrompt string, mcpClient domain.MCPClient) {
	h := handler.NewOllamaHandler(ollamaUrl, model, systemPrompt, mcpClient)

	router := e.Group("/api/v1/ollama")
	router.GET("/chat", h.Stream)
}
