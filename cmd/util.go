package main

import (
	"encoding/json"

	"github.com/openai/openai-go"
)

func extractContent(message openai.ChatCompletionMessageParamUnion) string {
	// Convert to JSON to extract content
	jsonData, err := json.Marshal(message)
	if err != nil {
		return ""
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return ""
	}

	if content, ok := parsed["content"].(string); ok {
		return content
	}

	return ""
}

func extractRole(message openai.ChatCompletionMessageParamUnion) string {
	// Convert to JSON to extract role
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "user"
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return "user"
	}

	if role, ok := parsed["role"].(string); ok {
		return role
	}

	return "user"
}

func extractContentFromMessage(message openai.ChatCompletionMessageParamUnion) string {
	return extractContent(message)
}

func extractRoleFromMessage(message openai.ChatCompletionMessageParamUnion) string {
	return extractRole(message)
}
