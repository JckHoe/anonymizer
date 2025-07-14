package main

import (
	"context"
	"encoding/json"
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
	log.Printf("Anonymizer is using model: %s\n", model)

	return &Anonymizer{
		systemPrompt: string(systemPrompt),
		model:        model,
		client:       client,
	}
}

func (an *Anonymizer) Anonymize(ctx context.Context, input string) map[string][]string {

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

	resp := completion.Choices[0].Message.Content

	// Unmarshal the response into a map array

	anonymizedData := make(map[string][]string)
	if err := json.Unmarshal([]byte(resp), &anonymizedData); err != nil {
		log.Fatalf("Failed to unmarshal anonymized data: %v", err)
	}
	return anonymizedData
}
