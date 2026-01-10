package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/metalpoch/local-synapse/internal/infrastructure/cache"
	"github.com/metalpoch/local-synapse/internal/pkg/config"
	"github.com/metalpoch/local-synapse/internal/routes"
)

var jwtSecret string
var port string
var ollamaModel string
var ollamaUrl string
var ollamaSystemPrompt string
var valkeyAddress string

func init() {
	if err := config.ApiEnviroment(&port, &jwtSecret, &ollamaUrl, &ollamaModel, &ollamaSystemPrompt, &valkeyAddress); err != nil {
		panic(err)
	}

}

func main() {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	routes.Init(&routes.Config{
		Echo:               e,
		Cache:              cache.NewValkeyClient(valkeyAddress),
		Secret:             jwtSecret,
		OllamaUrl:          ollamaUrl,
		OllamaModel:        ollamaModel,
		OllamaSystemPrompt: ollamaSystemPrompt,
	})

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
