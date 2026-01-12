package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/handler"
	"github.com/metalpoch/local-synapse/internal/infrastructure/cache"
	"github.com/metalpoch/local-synapse/internal/middleware"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
	"github.com/metalpoch/local-synapse/internal/repository"
	"github.com/metalpoch/local-synapse/internal/usecase/user"
)

func SetupOllamaRouter(
	e *echo.Echo,
	ollamaUrl, model, systemPrompt string,
	mcpClient domain.MCPClient,
	am authentication.AuthManager,
	ur repository.UserRepository,
	cr repository.ConversationRepository,
	cc cache.ConversationCache,
) {
	h := handler.NewOllamaHandler(
		ollamaUrl,
		model,
		systemPrompt,
		mcpClient,
		user.NewGetUser(ur),
		cr,
		cc,
	)

	router := e.Group("/api/v1/ollama", middleware.AuthMiddleware(am))
	router.GET("/chat", h.Stream)
	router.GET("/history", h.History)
}
