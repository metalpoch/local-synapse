package main

import (
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

	e.Logger.Fatal(e.Start(port))

}
