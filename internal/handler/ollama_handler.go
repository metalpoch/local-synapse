package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/dto"
)

type ollamaHandler struct {
	OllamaURL    string
	Model        string
	SystemPrompt string
}

func NewOllamaHandler(url, model, systemPrompt string) *ollamaHandler {
	return &ollamaHandler{url, model, systemPrompt}
}

func (hdlr *ollamaHandler) Stream(c echo.Context) error {
	userPrompt := c.QueryParam("prompt")
	if userPrompt == "" {
		return c.String(http.StatusBadRequest, "Query parameter 'prompt' is required")
	}

	format := c.QueryParam("format")
	isPlain := format == "plain"

	res := c.Response()
	if isPlain {
		res.Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	} else {
		res.Header().Set(echo.HeaderContentType, "text/event-stream")
	}
	res.Header().Set(echo.HeaderCacheControl, "no-cache")
	res.Header().Set(echo.HeaderConnection, "keep-alive")
	res.Header().Set("X-Content-Type-Options", "nosniff")
	res.Header().Set("X-Accel-Buffering", "no")
	res.WriteHeader(http.StatusOK)

	messages := []dto.OllamaChatMessage{
		{Role: "system", Content: hdlr.SystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	body, err := json.Marshal(dto.OllamaChatRequest{
		Model:    hdlr.Model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(c.Request().Context(), "POST", hdlr.OllamaURL+"/api/chat", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Failed to connect to Ollama")
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if isPlain {
			var oRes dto.OllamaChatResponse
			if err := json.Unmarshal([]byte(line), &oRes); err == nil {
				if _, err := fmt.Fprint(res.Writer, oRes.Message.Content); err != nil {
					return err
				}
			}
		} else {
			if _, err := fmt.Fprintf(res.Writer, "data: %s\n\n", line); err != nil {
				return err
			}
		}
		res.Flush()
	}

	return scanner.Err()
}
