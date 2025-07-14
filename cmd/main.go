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

func main() {
	ctx := context.Background()

	anonymizer := NewAnonymizer()
	client := NewOpenAIClient()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your messages (type 'quit' to exit):")

	var messages []openai.ChatCompletionMessageParamUnion
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

		// Test anonymizer
		anonymized := anonymizer.Anonymize(ctx, input)
		fmt.Printf("Anonymized Input: %s\n", anonymized)

		messages = append(messages, openai.UserMessage(input))

		completion, err := client.CreateChatCompletion(ctx, messages, "openai/gpt-4.1-mini")
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		response := completion.Choices[0].Message.Content
		fmt.Printf("Response: %s\n\n", response)

		// Add assistant response to history
		messages = append(messages, openai.AssistantMessage(response))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v\n", err)
	}
}
