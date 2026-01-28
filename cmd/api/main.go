package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	mcpclient "github.com/metalpoch/local-synapse/internal/infrastructure/mcp_client"
	"github.com/metalpoch/local-synapse/internal/pkg/config"
	"github.com/metalpoch/local-synapse/internal/router"
)

var (
	port               string
	ollamaModel        string
	ollamaUrl          string
	ollamaSystemPrompt string
	mcpClient          mcpclient.MCPClient
)

func init() {
	err := config.ApiEnviroment(&port, &ollamaUrl, &ollamaModel, &ollamaSystemPrompt)
	if err != nil {
		panic(err)
	}

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

	// Register all application routes
	router.SetupSystemRouter(e)
	router.SetupOllamaRouter(
		e,
		ollamaUrl,
		ollamaModel,
		ollamaSystemPrompt,
		mcpClient,
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
