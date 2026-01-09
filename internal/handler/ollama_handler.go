package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/dto"
)

type ollamaHandler struct {
	OllamaURL    string
	Model        string
	SystemPrompt string
	mcpClient    domain.MCPClient
}

func NewOllamaHandler(url, model, systemPrompt string, mcpClient domain.MCPClient) *ollamaHandler {
	return &ollamaHandler{url, model, systemPrompt, mcpClient}
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

	// Broad context
	ctx := c.Request().Context()

	// 1. Get Tools from MCP
	var tools []dto.Tool
	if hdlr.mcpClient != nil {
		mcpTools, err := hdlr.mcpClient.ListTools(ctx)
		if err == nil && len(mcpTools) > 0 {
			for _, t := range mcpTools {
				tools = append(tools, dto.Tool{
					Type: "function",
					Function: dto.ToolFunction{
						Name:        t.Name,
						Description: t.Description,
						Parameters:  t.InputSchema,
					},
				})
			}
		}
	}

	// 2. Initial Chat History
	messages := []dto.OllamaChatMessage{
		{Role: "system", Content: hdlr.SystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	// 3. First Request (No streaming to detect tools safely)
	req1 := dto.OllamaChatRequest{
		Model:    hdlr.Model,
		Messages: messages,
		Stream:   false, // Disable streaming for the decision phase
		Tools:    tools,
	}

	resp1, err := hdlr.callOllama(ctx, req1)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Failed to connect to Ollama")
	}

	// 4. Check for Tool Calls
	if len(resp1.Message.ToolCalls) > 0 {
		// Append assistant's "call" message
		messages = append(messages, dto.OllamaChatMessage{
			Role:      resp1.Message.Role, // usually "assistant"
			Content:   resp1.Message.Content,
			ToolCalls: resp1.Message.ToolCalls,
		})

		// Execute tools
		for _, tc := range resp1.Message.ToolCalls {
			// Notify user we are executing a tool? For now, silence (or maybe send a comment in plain text/SSE event)
			
			result, err := hdlr.mcpClient.CallTool(ctx, tc.Function.Name, tc.Function.Arguments)
			content := ""
			if err != nil {
				content = fmt.Sprintf("Error executing tool: %v", err)
			} else {
				// Convert result to string (Text or JSON)
				if len(result.Content) > 0 {
					for _, c := range result.Content {
						if tc, ok := c.(mcp.TextContent); ok {
							content += tc.Text + " "
						} else {
							b, _ := json.Marshal(c)
							content += string(b) + " "
						}
					}
				}
			}

			// Report back to Ollama
			messages = append(messages, dto.OllamaChatMessage{
				Role:    "tool",
				Content: content,
			})
		}

		// 5. Final Request (Streaming)
		req2 := dto.OllamaChatRequest{
			Model:    hdlr.Model,
			Messages: messages,
			Stream:   true,
			Tools:    tools, // Keep tools enabled just in case (multi-turn)
		}
		
		return hdlr.streamOllama(ctx, req2, res, isPlain)
	}

	// 5b. No Tools called - Stream the content we already got? or Stream request again?
	// Since we set Stream=false, we have the full content in resp1.Message.Content.
	// We can simply write it out.
	if isPlain {
		if _, err := fmt.Fprint(res.Writer, resp1.Message.Content); err != nil {
			return err
		}
	} else {
		// Mock SSE stream for the single response
		chunkReq := dto.OllamaChatResponse{}
		chunkReq.Message.Content = resp1.Message.Content
		chunkReq.Done = true
		
		chunkLine, _ := json.Marshal(chunkReq)
		if _, err := fmt.Fprintf(res.Writer, "data: %s\n\n", chunkLine); err != nil {
			return err
		}
	}
	res.Flush()

	return nil
}

// Helper to call Ollama (Non-streaming)
func (hdlr *ollamaHandler) callOllama(ctx context.Context, reqBody dto.OllamaChatRequest) (*dto.OllamaChatResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", hdlr.OllamaURL+"/api/chat", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status: %d", resp.StatusCode)
	}

	var oRes dto.OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&oRes); err != nil {
		return nil, err
	}
	oRes.Message.Role = "assistant" // Ensure role is set
	return &oRes, nil
}

// Helper to stream Ollama (Streaming)
func (hdlr *ollamaHandler) streamOllama(ctx context.Context, reqBody dto.OllamaChatRequest, res *echo.Response, isPlain bool) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", hdlr.OllamaURL+"/api/chat", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
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
