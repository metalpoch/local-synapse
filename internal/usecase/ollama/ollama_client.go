package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/metalpoch/local-synapse/internal/dto"
)

// OllamaClient handles HTTP communication with Ollama API
type OllamaClient struct {
	baseURL string
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{baseURL: baseURL}
}

// StreamChatRequest sends a chat request to Ollama and streams the response
// onChunk is called for each chunk received from Ollama
func (c *OllamaClient) StreamChatRequest(
	ctx context.Context,
	request dto.OllamaChatRequest,
	onChunk func(dto.OllamaChatResponse) error,
) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var chatResp dto.OllamaChatResponse
		if err := json.Unmarshal([]byte(line), &chatResp); err != nil {
			// Log but continue - some lines might be malformed
			continue
		}

		if err := onChunk(chatResp); err != nil {
			return fmt.Errorf("chunk handler error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}
