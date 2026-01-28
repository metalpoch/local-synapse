package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func ApiEnviroment(port, ollamaUrl, ollamaModel, ollamaSystemPrompt *string) error {
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

	p := os.Getenv("PORT")
	if p == "" {
		return errors.New("error 'PORT' environment variable required.")
	}

	if _, err := strconv.Atoi(p); err != nil {
		return fmt.Errorf("error 'PORT' must be a valid number: %v", err)
	}

	*port = ":" + p
	*ollamaUrl = ou
	*ollamaModel = om
	*ollamaSystemPrompt = osp

	return nil
}
