package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/handler"
)

func NewOllamaRoutes(server *echo.Echo, ollamaUrl, model, systemPrompt string, mcpClient domain.MCPClient) {
	hdlr := handler.NewOllamaHandler(ollamaUrl, model, systemPrompt, mcpClient)

	route := server.Group("/api/v1/ollama")
	route.GET("/chat", hdlr.Stream)
}
