package services

import (
	"os"
	"sync"

	"github.com/sashabaranov/go-openai"
)

var DefaultOpenAIClient = sync.OnceValue(func() *openai.Client {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY is not set")
	}
	return openai.NewClient(apiKey)
})
