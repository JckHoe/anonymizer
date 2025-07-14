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
	client := NewOpenAIClient()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your messages (type 'quit' to exit):")

	var messages []openai.ChatCompletionMessageParamUnion
	allAnonymizedData := make(AnonymizedData)

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
		allAnonymizedData.Merge(anonymized)
		fmt.Printf("%sAnonymized Data: %s%s\n", ColorYellow, anonymized, ColorReset)
		fmt.Printf("%sAll Anonymized Data: %s%s\n", ColorCyan, allAnonymizedData, ColorReset)

		messages = append(messages, openai.UserMessage(input))

		completion, err := client.CreateChatCompletion(ctx, messages, "openai/gpt-4.1-mini")
		if err != nil {
			log.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}

		response := completion.Choices[0].Message.Content
		fmt.Printf("%sResponse: %s%s\n\n", ColorGreen, response, ColorReset)

		// Add assistant response to history
		messages = append(messages, openai.AssistantMessage(response))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v\n", err)
	}
}
