package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)
func ApiEnviroment(port, jwtSecret, ollamaUrl, ollamaModel, ollamaSystemPrompt, valkeyAddress *string) error {
	js := os.Getenv("JWT_SECRET")
	if js == "" {
		return errors.New("error 'JWT_SECRET' environment variable required.")
	}

	ou := os.Getenv("OLLAMA_URL")
	if ou == "" {
		return errors.New("error 'OLLAMA_URL' environment variable required.")
	}

	om := os.Getenv("OLLAMA_MODEL")
	if om == "" {
		return errors.New("error 'OLLAMA_MODEL' environment variable required.")
	}

	osp := os.Getenv("OLLAMA_SYSTEM_PROMPT")
	if osp == "" {
		return errors.New("error 'OLLAMA_SYSTEM_PROMPT' environment variable required.")
	}

	va := os.Getenv("VALKEY_ADDRESS")
	if va == "" {
		va = "127.0.0.1:6379" // Default for local dev
	}

	p := os.Getenv("PORT")
	if p == "" {
		return errors.New("error 'PORT' environment variable required.")
	}

	if _, err := strconv.Atoi(p); err != nil {
		return fmt.Errorf("error 'PORT' must be a valid number: %v", err)
	}

	*port = ":" + p
	*jwtSecret = js
	*ollamaUrl = ou
	*ollamaModel = om
	*ollamaSystemPrompt = osp
	*valkeyAddress = va

	return nil
}
