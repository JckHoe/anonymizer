package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

type AnonymizedData map[string][]string

func (ad AnonymizedData) Merge(other AnonymizedData) {
	for key, values := range other {
		ad[key] = append(ad[key], values...)
	}
}

func (an *Anonymizer) Anonymize(ctx context.Context, input string) AnonymizedData {

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
	anonymizedData := make(AnonymizedData)
	if err := json.Unmarshal([]byte(resp), &anonymizedData); err != nil {
		log.Fatalf("Failed to unmarshal anonymized data: %v", err)
	}
	return anonymizedData
}

func (an *Anonymizer) AnonymizeMessages(
	ctx context.Context,
	messages []openai.ChatCompletionMessage,
	anonymizedData AnonymizedData,
) (*openai.ChatCompletion, error) {

	// Create anonymized versions of messages
	anonymizedMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))

	// Replace values in each message with key + number format
	for i, message := range messages {
		content := message.Content
		anonymizedContent := content

		// Replace each value with key + number format
		for key, values := range anonymizedData {
			for j, value := range values {
				replacement := fmt.Sprintf("[%s %d]", key, j+1)
				anonymizedContent = strings.ReplaceAll(anonymizedContent, value, replacement)
			}
		}

		// Create new message with anonymized content based on role
		switch message.Role {
		case "user":
			anonymizedMessages[i] = openai.UserMessage(anonymizedContent)
		case "assistant":
			anonymizedMessages[i] = openai.AssistantMessage(anonymizedContent)
		case "system":
			anonymizedMessages[i] = openai.SystemMessage(anonymizedContent)
		default:
			anonymizedMessages[i] = openai.UserMessage(anonymizedContent)
		}
	}

	// Send anonymized messages to OpenAI
	return an.client.CreateChatCompletion(ctx, anonymizedMessages, an.model)
}
