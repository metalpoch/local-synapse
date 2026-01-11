package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/middleware"
	"github.com/metalpoch/local-synapse/internal/usecase/ollama"
	"github.com/metalpoch/local-synapse/internal/usecase/user"
)

type ollamaHandler struct {
	streamChatUC *ollama.StreamChatUsecase
	getUser      *user.GetUser
}

func NewOllamaHandler(url, model, systemPrompt string, mcpClient domain.MCPClient, gu *user.GetUser) *ollamaHandler {
	return &ollamaHandler{
		ollama.NewStreamChatUsecase(url, model, systemPrompt, mcpClient),
		gu,
	}
}

func (hdlr *ollamaHandler) Stream(c echo.Context) error {
	userID, _ := middleware.GetUserID(c)

	userPrompt := c.QueryParam("prompt")
	if userPrompt == "" {
		return c.String(http.StatusBadRequest, "Query parameter 'prompt' is required")
	}

	format := c.QueryParam("format")
	isPlain := format == "plain"

	// Configure response headers
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

	// Define chunk handler for streaming
	onChunk := func(chunk dto.OllamaChatResponse) error {
		if isPlain {
			if chunk.Message.Thinking != "" {
				if _, err := fmt.Fprintf(res.Writer, "[Pensando: %s]\n", chunk.Message.Thinking); err != nil {
					return err
				}
			}
			if chunk.Message.Content != "" {
				if _, err := fmt.Fprint(res.Writer, chunk.Message.Content); err != nil {
					return err
				}
			}
		} else {
			// SSE format - only send chunks with actual content or thinking
			if chunk.Message.Content == "" && chunk.Message.Thinking == "" && len(chunk.Message.ToolCalls) == 0 {
				// Skip empty chunks
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

	user, err := hdlr.getUser.Execute(userID)
if userPrompt == "" {
	return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return hdlr.streamChatUC.StreamChat(ctx, user.Name, userPrompt, onChunk)
}
