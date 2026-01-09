package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
)

func NewOllamaRoutes(server *echo.Echo, ollamaUrl, model, systemPrompt string) {
	hdlr := handler.NewOllamaHandler(ollamaUrl, model, systemPrompt)

	route := server.Group("/api/v1/ollama")
	route.GET("/chat", hdlr.Stream)
}
