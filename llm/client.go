package llm

import (
	"context"
	"fmt"
	"os"

	cerebras "github.com/rizome-dev/go-cerebras/pkg"
)

type LLMClient struct {
	client *cerebras.Client
	model  string
	temp   float64
}

func NewLLMClient(model string, temperature float64) (*LLMClient, error) {
	apiKey := os.Getenv("CEREBRAS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CEREBRAS_API_KEY not set")
	}

	client := cerebras.NewClient(apiKey)

	return &LLMClient{
		client: client,
		model:  model,
		temp:   temperature,
	}, nil
}

func (c *LLMClient) Generate(prompt string) (string, error) {
	temp := float32(c.temp)
	resp, err := c.client.Chat.Create(context.Background(), &cerebras.ChatCompletionRequest{
		Model: c.model,
		Messages: []cerebras.Message{
			cerebras.NewUserMessage(prompt),
		},
		MaxCompletionTokens: cerebras.Int(256),
		Temperature:         &temp,
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *LLMClient) GenerateWithSystem(system, user string) (string, error) {
	temp := float32(c.temp)
	resp, err := c.client.Chat.Create(context.Background(), &cerebras.ChatCompletionRequest{
		Model: c.model,
		Messages: []cerebras.Message{
			cerebras.NewSystemMessage(system),
			cerebras.NewUserMessage(user),
		},
		MaxCompletionTokens: cerebras.Int(256),
		Temperature:         &temp,
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return resp.Choices[0].Message.Content, nil
}
