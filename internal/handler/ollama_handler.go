package handler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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
			log.Printf("[MCP] Discovered %d tools", len(mcpTools))
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
		} else if err != nil {
			log.Printf("[MCP] Error listing tools: %v", err)
		}
	}

	// 2. Initial Chat History
	messages := []dto.OllamaChatMessage{
		{Role: "system", Content: hdlr.SystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	// 3. First Request (Stream = true)
	req1 := dto.OllamaChatRequest{
		Model:    hdlr.Model,
		Messages: messages,
		Stream:   true,
		Tools:    tools,
	}

	log.Printf("[MCP] Sending initial request to Ollama (Streaming mode)")

	body, err := json.Marshal(req1)
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

	var fullContent string
	var allToolCalls []dto.ToolCall

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var oRes dto.OllamaChatResponse
		if err := json.Unmarshal([]byte(line), &oRes); err != nil {
			log.Printf("Error unmarshalling stream line: %v", err)
			continue
		}

		// Accumulate content
		fullContent += oRes.Message.Content
		// Accumulate tool calls (if any)
		if len(oRes.Message.ToolCalls) > 0 {
			allToolCalls = append(allToolCalls, oRes.Message.ToolCalls...)
		}

		// Stream to user immediately
		if isPlain {
			if _, err := fmt.Fprint(res.Writer, oRes.Message.Content); err != nil {
				return err
			}
		} else {
			// Forward the exact line from Ollama
			if _, err := fmt.Fprintf(res.Writer, "data: %s\n\n", line); err != nil {
				return err
			}
		}
		res.Flush()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// 4. Check for Tool Calls
	if len(allToolCalls) > 0 {
		log.Printf("[MCP] Ollama requested %d tools", len(allToolCalls))
		// Append assistant's "call" message
		messages = append(messages, dto.OllamaChatMessage{
			Role:      "assistant",
			Content:   fullContent,
			ToolCalls: allToolCalls,
		})

		// Execute tools
		for _, tc := range allToolCalls {
			log.Printf("[MCP] Executing tool: %s", tc.Function.Name)

			result, err := hdlr.mcpClient.CallTool(ctx, tc.Function.Name, tc.Function.Arguments)
			content := ""
			if err != nil {
				log.Printf("[MCP] Tool execution failed: %v", err)
				content = fmt.Sprintf("Error executing tool: %v", err)
			} else {
				log.Printf("[MCP] Tool execution successful")
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
		log.Printf("[MCP] Sending final request with tool results")
		req2 := dto.OllamaChatRequest{
			Model:    hdlr.Model,
			Messages: messages,
			Stream:   true,
			Tools:    tools, // Keep tools enabled just in case (multi-turn)
		}

		return hdlr.streamOllama(ctx, req2, res, isPlain)
	}

	return nil
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
