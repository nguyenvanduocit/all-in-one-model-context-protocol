package services

import (
	"os"
	"sync"

	"github.com/sashabaranov/go-openai"
)

var DefaultDeepseekClient = sync.OnceValue(func() *openai.Client {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		panic("DEEPSEEK_API_KEY environment variable must be set")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com"

	client := openai.NewClientWithConfig(config)
	return client
})