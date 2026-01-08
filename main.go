package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
)

const (
	MODEL      = "qwen3:8b"
	DEFAULT_URL = "http://host.containers.internal:11434"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type OllamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func getOllamaURL() string {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		return DEFAULT_URL
	}
	return url
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	userPrompt := r.URL.Query().Get("prompt")
	if userPrompt == "" {
		http.Error(w, "Query parameter 'prompt' is required", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	isPlain := format == "plain"

	systemPrompt := os.Getenv("SYSTEM_PROMPT")
	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	if isPlain {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "text/event-stream")
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	body, err := json.Marshal(OllamaChatRequest{
		Model:    MODEL,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		http.Error(w, "Failed to create request body", http.StatusInternalServerError)
		log.Printf("Error marshaling request body: %v", err)
		return
	}

	ollamaURL := getOllamaURL()
	req, err := http.NewRequestWithContext(r.Context(), "POST", ollamaURL+"/api/chat", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to connect to Ollama service", http.StatusServiceUnavailable)
		log.Printf("Error connecting to Ollama at %s: %v", ollamaURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Ollama API returned status: %s", resp.Status), http.StatusBadGateway)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if isPlain {
			var res OllamaChatResponse
			if err := json.Unmarshal([]byte(line), &res); err == nil {
				fmt.Fprint(w, res.Message.Content)
			}
		} else {
			if _, err := fmt.Fprintf(w, "data: %s\n\n", line); err != nil {
				log.Printf("Error writing to client: %v", err)
				return
			}
		}
		flusher.Flush()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading Ollama response: %v", err)
	}
}


func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/chat", streamHandler)
	log.Printf("Proxy iniciado en puerto %s, apuntando a %s", port, getOllamaURL())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

