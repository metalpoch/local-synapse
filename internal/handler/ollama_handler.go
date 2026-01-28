package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/usecase/ollama"
)

type ollamaHandler struct {
	chatUC *ollama.StreamChatUsecase
}

func NewOllamaHandler(chatUC *ollama.StreamChatUsecase) *ollamaHandler {
	return &ollamaHandler{chatUC}
}

func (h *ollamaHandler) Stream(c echo.Context) error {
	userPrompt := c.QueryParam("prompt")
	if userPrompt == "" {
		return c.String(http.StatusBadRequest, "Query parameter 'prompt' is required")
	}

	format := c.QueryParam("format")
	isPlain := format == "plain"

	// Set up streaming response headers
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

	ctx := c.Request().Context()

	// Process each response chunk from the LLM
	onChunk := func(chunk dto.OllamaChatResponse) error {
		if isPlain {
			if chunk.Message.Thinking != "" {
				if _, err := fmt.Fprintf(res.Writer, "[Thinking: %s]\n", chunk.Message.Thinking); err != nil {
					return err
				}
			}
			if chunk.Message.Content != "" {
				if _, err := fmt.Fprint(res.Writer, chunk.Message.Content); err != nil {
					return err
				}
			}
		} else {
			// Skip chunks without relevant updates for SSE
			if chunk.Message.Content == "" && chunk.Message.Thinking == "" && len(chunk.Message.ToolCalls) == 0 {
				return nil
			}

			jsonData, err := json.Marshal(chunk)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(res.Writer, "data: %s\n\n", jsonData); err != nil {
				return err
			}
		}
		res.Flush()
		return nil
	}

	return h.chatUC.Execute(ctx, userPrompt, onChunk)
}
