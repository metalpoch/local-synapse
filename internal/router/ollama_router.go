package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
	"github.com/metalpoch/local-synapse/internal/infrastructure/cache"
	mcpclient "github.com/metalpoch/local-synapse/internal/infrastructure/mcp_client"
	"github.com/metalpoch/local-synapse/internal/middleware"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
	"github.com/metalpoch/local-synapse/internal/repository"
	"github.com/metalpoch/local-synapse/internal/usecase/ollama"
	"github.com/metalpoch/local-synapse/internal/usecase/user"
)

func SetupOllamaRouter(
	e *echo.Echo,
	ollamaUrl, model, systemPrompt string,
	mcpClient mcpclient.MCPClient,
	am authentication.AuthManager,
	ur repository.UserRepository,
	cr repository.ConversationRepository,
	cc cache.ConversationCache,
) {
	h := handler.NewOllamaHandler(
		ollama.NewStreamChatUsecase(ollamaUrl, model, systemPrompt, mcpClient, cr, cc),
		ollama.NewGetChatHistory(cr),
		ollama.NewListConversations(cr),
		ollama.NewCreateConversation(cr),
		ollama.NewDeleteConversation(cr),
		ollama.NewRenameConversation(cr),
		user.NewGetUser(ur),
	)

	router := e.Group("/api/v1/ollama", middleware.AuthMiddleware(am))
	router.GET("/chat", h.Stream)
	router.GET("/history", h.History)
	router.GET("/conversations", h.ListConversations)
	router.POST("/conversations", h.CreateConversation)
	router.DELETE("/conversations/:id", h.DeleteConversation)
	router.PUT("/conversations/:id/title", h.RenameConversation)
}
