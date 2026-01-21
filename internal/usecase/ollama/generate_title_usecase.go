package ollama

import (
	"context"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/infrastructure/ollama"
)

type GenerateTitleUsecase struct {
	ollamaClient *ollama_infra.OllamaClient
	model string
}

func NewGenerateTitleUsecase(baseURL, model string) *GenerateTitleUsecase {
	return &GenerateTitleUsecase{
		ollamaClient: ollama_infra.NewOllamaClient(baseURL),
		model: model,
	}
}

func (uc *GenerateTitleUsecase) Execute(ctx context.Context, userPrompt string) (string, error) {
	request := dto.OllamaChatRequest{
		Model: uc.model,
		Messages: []dto.OllamaChatMessage{
			{
				Role:    "system",
				Content: "You are a title generator. Generate a 3-5 word title in Spanish for the user's message. Reply ONLY with the title without any symbols, quotes or extra text.",
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	resp, err := uc.ollamaClient.ChatRequest(ctx, request)
	if err != nil {
		return "", err
	}

	return resp.Message.Content, nil
}
