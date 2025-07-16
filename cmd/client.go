package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type LLMClient interface {
	CreateChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error)
}

var _ LLMClient = (*OpenAIClient)(nil)
var _ LLMClient = (*OllamaClient)(nil)
var _ LLMClient = (*PIIDetectorClient)(nil)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient() *OpenAIClient {
	baseURL := "https://openrouter.ai/api/v1"
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}
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
	log.Printf("OpenAI API call took: %v\n", duration)
	return completion, err
}

type OllamaClient struct {
	client *openai.Client
}

func NewOllamaClient() *OllamaClient {
	baseURL := "http://localhost:11434/v1"
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
	log.Printf("Ollama API call took: %v\n", duration)
	return completion, err
}

type PIIDetectorClient struct {
	client    *http.Client
	serverURL string
}

func NewPIIDetectorClient(serverURL string) *PIIDetectorClient {
	return &PIIDetectorClient{
		client:    &http.Client{Timeout: 30 * time.Second},
		serverURL: serverURL,
	}
}

func (c *PIIDetectorClient) CreateChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, model string) (*openai.ChatCompletion, error) {
	start := time.Now()

	var inputText string
	for i := len(messages) - 1; i >= 0; i-- {
		role := extractRoleFromMessage(messages[i])
		if role == "user" {
			inputText = extractContentFromMessage(messages[i])
			break
		}
	}

	if inputText == "" {
		return nil, fmt.Errorf("no user message found to process")
	}

	requestBody := map[string]string{
		"input": inputText,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/infer", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Output any `json:"output"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	responseContent, err := c.convertPIIResultsToAnonymizationFormat(apiResponse.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PII results: %w", err)
	}

	completion := &openai.ChatCompletion{
		ID:      "pii-detector-" + fmt.Sprintf("%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Content: responseContent,
				},
			},
		},
	}

	duration := time.Since(start)
	log.Printf("PII Detector API call took: %v\n", duration)

	return completion, nil
}

func (c *PIIDetectorClient) convertPIIResultsToAnonymizationFormat(output interface{}) (string, error) {
	anonymizedData := map[string][]string{
		"Person":   []string{},
		"Location": []string{},
		"Company":  []string{},
		"Email":    []string{},
		"Phone":    []string{},
	}

	switch v := output.(type) {
	case map[string]any:
		for category, entities := range v {
			if entitiesList, ok := entities.([]interface{}); ok {
				var stringList []string
				for _, entity := range entitiesList {
					if str, ok := entity.(string); ok {
						stringList = append(stringList, str)
					}
				}
				mappedCategory := c.mapPIICategory(category)
				if mappedCategory != "" {
					anonymizedData[mappedCategory] = stringList
				}
			}
		}
	case []any:
		for _, entity := range v {
			if entityMap, ok := entity.(map[string]interface{}); ok {
				if entityType, ok := entityMap["type"].(string); ok {
					if entityValue, ok := entityMap["value"].(string); ok {
						mappedCategory := c.mapPIICategory(entityType)
						if mappedCategory != "" {
							anonymizedData[mappedCategory] = append(anonymizedData[mappedCategory], entityValue)
						}
					}
				}
			}
		}
	case string:
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			return c.convertPIIResultsToAnonymizationFormat(parsed)
		}
	}

	jsonBytes, err := json.Marshal(anonymizedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal anonymized data: %w", err)
	}

	return string(jsonBytes), nil
}

func (c *PIIDetectorClient) mapPIICategory(category string) string {
	lowerCategory := strings.ToLower(category)
	switch {
	case strings.Contains(lowerCategory, "person") || strings.Contains(lowerCategory, "name"):
		return "Person"
	case strings.Contains(lowerCategory, "location") || strings.Contains(lowerCategory, "address"):
		return "Location"
	case strings.Contains(lowerCategory, "company") || strings.Contains(lowerCategory, "organization"):
		return "Company"
	case strings.Contains(lowerCategory, "email"):
		return "Email"
	case strings.Contains(lowerCategory, "phone"):
		return "Phone"
	default:
		return ""
	}
}

func CreateLLMClient() (LLMClient, string, error) {
	model := os.Getenv("MODEL_NAME")
	if model == "" {
		// model = "llama3.2:3b"
		model = "openai/gpt-4.1-mini"
	}

	// Check if PII detector should be used
	if model == "pii-detector" || strings.HasPrefix(model, "pii-detector") {
		piiDetectorUrl := os.Getenv("PII_DETECTOR_URL")
		return NewPIIDetectorClient(piiDetectorUrl), model, nil
	}

	// Determine if this is an OpenAI model based on model name
	isOpenAIModel := model == "openai/gpt-4.1-mini" ||
		(len(model) > 6 && model[:6] == "openai") ||
		(len(model) > 3 && model[:3] == "gpt")

	if isOpenAIModel {
		return NewOpenAIClient(), model, nil
	} else {
		return NewOllamaClient(), model, nil
	}
}
