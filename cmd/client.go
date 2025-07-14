package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type LLMClient interface {
	CreateChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error)
}

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(apiKey, baseURL string) *OpenAIClient {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	return &OpenAIClient{client: &client}
}

func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
	})
	return completion, err
}

type OllamaClient struct {
	client *openai.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	client := openai.NewClient(
		option.WithAPIKey(""),
		option.WithBaseURL(baseURL),
	)
	return &OllamaClient{client: &client}
}

func (c *OllamaClient) CreateChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
	})
	return completion, err
}

func CreateLLMClient() (LLMClient, string, error) {
	model := os.Getenv("MODEL_NAME")
	
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		if model == "" {
			model = "openai/gpt-4.1-mini"
		}
		return NewOpenAIClient(os.Getenv("OPENROUTER_API_KEY"), "https://openrouter.ai/api/v1"), model, nil
	}
	
	if model != "" && (model == "openai/gpt-4.1-mini" || model[:6] == "openai" || model[:3] == "gpt") {
		return nil, "", fmt.Errorf("OpenAI model '%s' selected but OPENROUTER_API_KEY is not set", model)
	}
	
	if model == "" {
		model = "llama3.1:8b"
	}
	return NewOllamaClient(os.Getenv("OLLAMA_BASE_URL")), model, nil
}