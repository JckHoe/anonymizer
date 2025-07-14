package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"
)

func main() {
	ctx := context.Background()

	client, model, err := CreateLLMClient()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Using model: %s\n", model)

	systemPrompt, err := os.ReadFile("system_prompt.tmpl")
	if err != nil {
		log.Fatal("Failed to read system prompt:", err)
	}

	completion, err := client.CreateChatCompletion(ctx, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(string(systemPrompt)),
		openai.UserMessage("I have a friend, Amy, who has a dog named Max. I have another field named Bob who has a cat named Luna."),
	}, model)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion.Choices[0].Message.Content)
}
