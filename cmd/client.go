package main

import (
	"context"
	"fmt"
	"os"
	"time"

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
	start := time.Now()
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
	})
	duration := time.Since(start)
	fmt.Printf("OpenAI API call took: %v\n", duration)
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
	start := time.Now()
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
	})
	duration := time.Since(start)
	fmt.Printf("Ollama API call took: %v\n", duration)
	return completion, err
}

func CreateLLMClient() (LLMClient, string, error) {
	model := os.Getenv("MODEL_NAME")
	if model == "" {
		model = "llama3.2:1b"
	}
	
	// Determine if this is an OpenAI model based on model name
	isOpenAIModel := model == "openai/gpt-4.1-mini" || 
		(len(model) > 6 && model[:6] == "openai") || 
		(len(model) > 3 && model[:3] == "gpt")
	
	if isOpenAIModel {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			return nil, "", fmt.Errorf("OpenAI model '%s' selected but OPENROUTER_API_KEY is not set", model)
		}
		return NewOpenAIClient(apiKey, "https://openrouter.ai/api/v1"), model, nil
	}
	
	// Use Ollama for all other models
	return NewOllamaClient(os.Getenv("OLLAMA_BASE_URL")), model, nil
}