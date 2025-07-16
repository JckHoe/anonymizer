package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openai/openai-go"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func main() {
	ctx := context.Background()

	anonymizer := NewAnonymizer()
	secretSelector, err := NewLLMSecretSelector()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your messages (type 'quit' to exit):")

	var messages []openai.ChatCompletionMessageParamUnion
	allSecrets := make(Secrets)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			break
		}

		if input == "" {
			continue
		}

		secrets, err := secretSelector.Select(ctx, input)
		if err != nil {
			log.Fatal(err)
		}
		allSecrets.Merge(secrets)
		fmt.Printf("%Secret Data: %s%s\n", ColorYellow, secrets, ColorReset)
		fmt.Printf("%sAll Secret Data: %s%s\n", ColorCyan, allSecrets, ColorReset)

		messages = append(messages, openai.UserMessage(input))

		// Print current messages content
		printMessages("Current Messages", messages, ColorBlue)

		// Use AnonymizeMessages with conversation history
		anonymizedMessages := anonymizer.AnonymizeMessages(ctx, messages, allSecrets)

		// Print anonymized messages content
		printMessages("Anonymized Messages", anonymizedMessages, ColorPurple)

		// Create OpenAI client and call with anonymized messages
		client := NewOpenAIClient()
		completion, err := client.CreateChatCompletion(ctx, anonymizedMessages, "openai/gpt-4.1-mini")
		if err != nil {
			log.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}

		response := completion.Choices[0].Message.Content
		fmt.Printf("Response: %s\n\n", response)

		// Deanonymize the response for demonstration
		responseMessage := []openai.ChatCompletionMessageParamUnion{
			openai.AssistantMessage(response),
		}

		// Debug: Print anonymized data to understand the mapping
		fmt.Printf("%sDebug - Anonymized Data for deanonymization: %s%s\n", ColorYellow, allSecrets, ColorReset)

		deanonymizedResponse := anonymizer.DeanonymizeMessages(ctx, responseMessage, allSecrets)
		printMessages("Deanonymized Response", deanonymizedResponse, ColorGreen)

		// Add deanonymized assistant response to history
		deanonymizedContent := extractContentFromMessage(deanonymizedResponse[0])
		messages = append(messages, openai.AssistantMessage(deanonymizedContent))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v\n", err)
	}
}

func printMessages(title string, messages []openai.ChatCompletionMessageParamUnion, color string) {
	fmt.Printf("%s%s:%s\n", color, title, ColorReset)
	for _, msg := range messages {
		content := extractContentFromMessage(msg)
		role := extractRoleFromMessage(msg)
		fmt.Printf("%s  [%s]: %s%s\n", color, role, content, ColorReset)
	}
}
