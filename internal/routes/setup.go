package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/valkey-io/valkey-go"
)

type Config struct {
	Echo               *echo.Echo
	Cache              *valkey.Client
	Secret             string
	OllamaUrl          string
	OllamaModel        string
	OllamaSystemPrompt string
}

func Init(cfg *Config) {
	NewSystemRoutes(cfg.Echo)
	NewOllamaRoutes(cfg.Echo, cfg.OllamaUrl, cfg.OllamaModel, cfg.OllamaSystemPrompt)
}
