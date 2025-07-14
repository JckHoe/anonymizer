package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	anonymizer := NewAnonymizer()
	result := anonymizer.Anonymize(ctx, "I have a friend, Amy, who has a dog named Max. I have another field named Bob who has a cat named Luna.")
	log.Println("Anonymized result:", result)
}
