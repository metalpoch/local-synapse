package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valkey-io/valkey-go"

	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/infrastructure/cache"
	"github.com/metalpoch/local-synapse/internal/infrastructure/database"
	mcpclient "github.com/metalpoch/local-synapse/internal/infrastructure/mcp_client"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
	"github.com/metalpoch/local-synapse/internal/pkg/config"
	"github.com/metalpoch/local-synapse/internal/repository"
	"github.com/metalpoch/local-synapse/internal/router"
)

var (
	jwtSecret          string
	port               string
	ollamaModel        string
	ollamaUrl          string
	ollamaSystemPrompt string
	valkeyAddress      string
	valkeyPassword     string
	SqliteAddr         string
	db                 *sql.DB
	vlk                valkey.Client
	mcpClient          domain.MCPClient
	accessTokenTTL     time.Duration = 15 * time.Minute
	refreshToken       time.Duration = 7 * 24 * time.Hour
)

func init() {
	err := config.ApiEnviroment(&port, &jwtSecret, &ollamaUrl, &ollamaModel, &ollamaSystemPrompt, &valkeyAddress, &valkeyPassword, &SqliteAddr)
	if err != nil {
		panic(err)
	}

	db = database.NewSqliteClient(SqliteAddr)
	vlk = cache.NewValkeyClient(valkeyAddress, valkeyPassword)
	mcpClient, err = mcpclient.NewStdioClient("./mcp")

	if err != nil {
		log.Printf("failed to create MCP client: %v", err)
	} else {
		if err := mcpClient.Initialize(context.Background()); err != nil {
			log.Printf("failed to initialize MCP client: %v", err)
		}
	}

}

func main() {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.Static("/assets", "public/assets")

	// Global authentication manager
	authManager := authentication.NewAuthManager([]byte(jwtSecret), vlk, accessTokenTTL, refreshToken)

	// Initialize persistence and cache layers
	userRepository := repository.NewUserRepo(db)
	conversationRepository := repository.NewConversationRepository(db)
	conversationCache := cache.NewConversationCache(vlk)

	// Register all application routes
	router.SetupSystemRouter(e, authManager)
	router.SetupAuthRouter(e, authManager, userRepository)
	router.SetupOllamaRouter(
		e,
		ollamaUrl,
		ollamaModel,
		ollamaSystemPrompt,
		mcpClient,
		authManager,
		userRepository,
		conversationRepository,
		conversationCache,
	)

	// Start the server in a background goroutine
	go func() {
		if err := e.Start(port); err != nil && err != http.ErrServerClosed {
			e.Logger.Errorf("server error: %v", err)
		}
	}()

	// Wait for termination signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	e.Logger.Info("closing the server securely...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
