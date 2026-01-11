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

	"github.com/metalpoch/local-synapse/internal/infrastructure/cache"
	mcpclient "github.com/metalpoch/local-synapse/internal/infrastructure/mcp_client"
	"github.com/metalpoch/local-synapse/internal/infrastructure/sqlite"
	"github.com/metalpoch/local-synapse/internal/pkg/config"
	"github.com/metalpoch/local-synapse/internal/repository"
	"github.com/metalpoch/local-synapse/internal/router"
)

var jwtSecret string
var port string
var ollamaModel string
var ollamaUrl string
var ollamaSystemPrompt string
var valkeyAddress string
var valkeyPassword string
var SqliteAddr string

var db *sql.DB
var vlk *valkey.Client

func init() {
	if err := config.ApiEnviroment(
		&port,
		&jwtSecret,
		&ollamaUrl,
		&ollamaModel,
		&ollamaSystemPrompt,
		&valkeyAddress,
		&valkeyPassword,
		&SqliteAddr,
	); err != nil {
		panic(err)
	}

	db = sqlite.NewSqliteClient(SqliteAddr)
	vlk = cache.NewValkeyClient(valkeyAddress, valkeyPassword)
}

func main() {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Initialize MCP Client
	mcpClient, err := mcpclient.NewStdioClient("./mcp")
	if err != nil {
		log.Printf("Failed to create MCP client: %v", err)
	} else {
		if err := mcpClient.Initialize(context.Background()); err != nil {
			log.Printf("Failed to initialize MCP client: %v", err)
		}
	}

	// Initialize repositories
	userRepository := repository.NewUserRepo(db)

	// Setup routes
	router.SetupSystemRouter(e)
	router.SetupAuthRouter(e, userRepository)
	router.SetupOllamaRouter(e, ollamaUrl, ollamaModel, ollamaSystemPrompt, mcpClient)

	go func() {
		if err := e.Start(port); err != nil && err != http.ErrServerClosed {
			e.Logger.Errorf("Error en el servidor: %v", err)
		}
	}()

	//  Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	e.Logger.Info("Cerrando el servidor de forma segura...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
