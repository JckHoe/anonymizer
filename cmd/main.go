package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	var baseUrl, apiKey string

	ctx := context.Background()
	model := os.Getenv("MODEL_NAME")
	if model == "" {
		if os.Getenv("OPENROUTER_API_KEY") != "" {
			model = "openai/gpt-4.1-mini"
		} else {
			model = "llama3.1:8b"
		}
	}
	fmt.Printf("Using model: %s\n", model)

	// Check if we should use OpenRouter based on model name
	if model == "openai/gpt-4.1-mini" || (os.Getenv("OPENROUTER_API_KEY") != "" && model != "llama3.1:8b") {
		// Use OpenRouter
		baseUrl = "https://openrouter.ai/api/v1"
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	} else {
		// Use Ollama
		baseUrl = os.Getenv("OLLAMA_BASE_URL")
		if baseUrl == "" {
			baseUrl = "http://localhost:11434/v1"
		}
		apiKey = ""
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseUrl),
	)

	systemPrompt, err := os.ReadFile("system_prompt.tmpl")
	if err != nil {
		log.Fatal("Failed to read system prompt:", err)
	}

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(string(systemPrompt)),
			openai.UserMessage("I have a friend, Amy, who is a software engineer. She is very talented and has worked on many projects. Can you tell me more about her?"),
		},
		Model: model,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion.Choices[0].Message.Content)
}
