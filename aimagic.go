package main

import (
	"context"
	"encoding/json"
	"github.com/ayush6624/go-chatgpt"
	"log"
)

type AiResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Choices   []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func AiMagic(request string) AiResponse {
	c, err := chatgpt.NewClient(config.ApiKeys.OpenaiApikey)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	res, err := c.Send(ctx, &chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT4,
		Messages: []chatgpt.ChatMessage{
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: request,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	choicesJSON, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	var aiResponse AiResponse
	if err := json.Unmarshal(choicesJSON, &aiResponse); err != nil {
		log.Fatal(err)
	}

	return aiResponse
}
