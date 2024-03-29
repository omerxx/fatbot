package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/getsentry/sentry-go"
	openai "github.com/sashabaranov/go-openai"
)

func GetAiResponse(labels []string) string {
	client := openai.NewClient(os.Getenv("OPENAI_APITOKEN"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Temperature: 1.2,
			Model:       openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("You are funny David Goggins. Write a response to a user after their workout, congratulating them for their effort and enoucraging them to continue working out, address this list of words in your response: %s. Keep it under 100 characters. End the message with emojis matching the words from the list.", labels),
				},
			},
		},
	)

	if err != nil {
		log.Errorf("ChatCompletion error: %v\n", err)
		sentry.CaptureException(err)
		return ""
	}
	return resp.Choices[0].Message.Content
}

func GetAiWelcomeResponse() string {
	client := openai.NewClient(os.Getenv("OPENAI_APITOKEN"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Temperature: 1.2,
			Model:       openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "You are funny David Goggins. Write a response to a user after their workout, welcoming them back. Keep it under 100 characters.",
				},
			},
		},
	)

	if err != nil {
		log.Errorf("ChatCompletion error: %v\n", err)
		sentry.CaptureException(err)
		return ""
	}
	return resp.Choices[0].Message.Content
}
