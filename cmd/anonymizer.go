package main

import (
	"context"
	"log"
	"os"

	"github.com/openai/openai-go"
)

type Anonymizer struct {
	systemPrompt string
	model        string
	client       LLMClient
}

func NewAnonymizer() *Anonymizer {
	systemPrompt, err := os.ReadFile("system_prompt.tmpl")
	if err != nil {
		log.Fatal("Failed to read system prompt:", err)
	}

	client, model, err := CreateLLMClient()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using model: %s\n", model)

	return &Anonymizer{
		systemPrompt: string(systemPrompt),
		model:        model,
		client:       client,
	}
}

func (an *Anonymizer) Anonymize(ctx context.Context, input string) string {

	systemPrompt, err := os.ReadFile("system_prompt.tmpl")
	if err != nil {
		log.Fatal("Failed to read system prompt:", err)
	}

	completion, err := an.client.CreateChatCompletion(ctx, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(string(systemPrompt)),
		openai.UserMessage(input),
	}, an.model)

	if err != nil {
		log.Fatal(err)
	}

	return completion.Choices[0].Message.Content
}
