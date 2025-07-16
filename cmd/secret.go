package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"
)

type Secrets map[string][]string

func (ad Secrets) Merge(other Secrets) {
	for key, values := range other {
		ad[key] = append(ad[key], values...)
	}
}

type SecretSelector interface {
	Select(context.Context, string) (Secrets, error)
}

type LLMSecretSelector struct {
	systemPrompt string
	model        string
	client       LLMClient
}

func NewLLMSecretSelector() (*LLMSecretSelector, error) {
	systemPrompt, err := os.ReadFile("system_prompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("Failed to read system prompt:", err)
	}

	client, model, err := CreateLLMClient()
	if err != nil {
		return nil, err
	}
	log.Printf("Anonymizer is using model: %s\n", model)

	return &LLMSecretSelector{
		systemPrompt: string(systemPrompt),
		model:        model,
		client:       client,
	}, nil
}

func (an *LLMSecretSelector) Select(ctx context.Context, input string) (Secrets, error) {

	systemPrompt, err := os.ReadFile("system_prompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("Failed to read system prompt:", err)
	}

	completion, err := an.client.CreateChatCompletion(ctx, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(string(systemPrompt)),
		openai.UserMessage(input),
	}, an.model)

	if err != nil {
		return nil, fmt.Errorf("Failed to create chat completion: %v", err)
	}

	resp := completion.Choices[0].Message.Content

	// Unmarshal the response into a map array
	secrets := make(Secrets)
	if err := json.Unmarshal([]byte(resp), &secrets); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal anonymized data: %v", err)
	}
	return secrets, nil
}
