package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/openai/openai-go"
)

type Anonymizer struct {
}

func NewAnonymizer() *Anonymizer {
	return &Anonymizer{}
}

func (an *Anonymizer) AnonymizeMessages(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	anonymizedData Secrets,
) []openai.ChatCompletionMessageParamUnion {

	// Create anonymized versions of messages
	anonymizedMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))

	// Replace values in each message with key + number format
	for i, message := range messages {
		// Extract actual content from the message
		content := extractContent(message)
		anonymizedContent := content

		// Replace each value with key + number format using word boundaries
		for key, values := range anonymizedData {
			for j, value := range values {
				replacement := fmt.Sprintf("[%s %d]", key, j+1)
				// Use regex with word boundaries to match whole words only
				pattern := `\b` + regexp.QuoteMeta(value) + `\b`
				re := regexp.MustCompile(pattern)
				anonymizedContent = re.ReplaceAllString(anonymizedContent, replacement)
			}
		}

		// Create new message with anonymized content based on role
		role := extractRole(message)
		switch role {
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

	// Return anonymized messages
	return anonymizedMessages
}

func (an *Anonymizer) DeanonymizeMessages(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	anonymizedData Secrets,
) []openai.ChatCompletionMessageParamUnion {

	// Create deanonymized versions of messages
	deanonymizedMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))

	// Replace anonymized placeholders back to original values
	for i, message := range messages {
		// Extract actual content from the message
		content := extractContent(message)
		deanonymizedContent := content

		// Replace each anonymized placeholder with original value
		for key, values := range anonymizedData {
			for j, value := range values {
				// Replace both bracketed and unbracketed versions
				bracketedPlaceholder := fmt.Sprintf("[%s %d]", key, j+1)
				unbracketedPlaceholder := fmt.Sprintf("%s %d", key, j+1)

				// log.Printf("Trying to replace '%s' and '%s' with '%s' in: %s", bracketedPlaceholder, unbracketedPlaceholder, value, deanonymizedContent)

				// Replace bracketed version
				pattern1 := regexp.QuoteMeta(bracketedPlaceholder)
				re1 := regexp.MustCompile(pattern1)
				// beforeReplace := deanonymizedContent
				deanonymizedContent = re1.ReplaceAllString(deanonymizedContent, value)

				// Replace unbracketed version
				pattern2 := regexp.QuoteMeta(unbracketedPlaceholder)
				re2 := regexp.MustCompile(pattern2)
				deanonymizedContent = re2.ReplaceAllString(deanonymizedContent, value)

				// if beforeReplace != deanonymizedContent {
				// 	log.Printf("Successfully replaced placeholder with '%s'", value)
				// }
			}
		}

		// Create new message with deanonymized content based on role
		role := extractRole(message)
		switch role {
		case "user":
			deanonymizedMessages[i] = openai.UserMessage(deanonymizedContent)
		case "assistant":
			deanonymizedMessages[i] = openai.AssistantMessage(deanonymizedContent)
		case "system":
			deanonymizedMessages[i] = openai.SystemMessage(deanonymizedContent)
		default:
			deanonymizedMessages[i] = openai.UserMessage(deanonymizedContent)
		}
	}

	// Return deanonymized messages
	return deanonymizedMessages
}
