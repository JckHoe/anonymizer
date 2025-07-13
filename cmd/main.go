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
	baseUrl := "https://openrouter.ai/api/v1"
	// get from env
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseUrl),
	)

	systemPrompt, err := os.ReadFile("system_prompt.tmpl")
	if err != nil {
		log.Fatal("Failed to read system prompt:", err)
	}

	ctx := context.Background()

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(string(systemPrompt)),
			openai.UserMessage("I have a friend, Amy, who is a software engineer. She is very talented and has worked on many projects. Can you tell me more about her?"),
		},
		Model: "openai/gpt-4.1-mini",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion.Choices[0].Message.Content)
}
