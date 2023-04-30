package openai

import (
	"context"
	"time"

	measure "github.com/portainer/portainer-ee/api/internal/time"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

// SendChatCompletionRequest sends a chat completion request to the OpenAI API.
// https://platform.openai.com/docs/guides/chat
func SendChatCompletionRequest(context context.Context, model string, prompt map[string][]string, apiKey string) (string, error) {
	client := openai.NewClient(apiKey)

	log.Debug().Msg("sending OpenAI chat completion request")
	defer measure.TrackTime(time.Now(), "OpenAI chat completion request")

	resp, err := client.CreateChatCompletion(
		context,
		openai.ChatCompletionRequest{
			Model:       model,
			MaxTokens:   2048,
			TopP:        0.10,
			Temperature: 0,
			Messages:    buildMessageSetWithMultipleUserMessages(prompt),
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("an error occured while sending OpenAI chat completion request")
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

// buildMessageSetWithMultipleUserMessages builds a message set for the OpenAI chat completion request.
// Each user message is sent as a separate message.
func buildMessageSetWithMultipleUserMessages(prompt map[string][]string) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{}
	for k, v := range prompt {
		for i := range v {
			msg := openai.ChatCompletionMessage{
				Role:    k,
				Content: v[i],
			}

			log.Debug().Str("Role", k).Str("Content", msg.Content).Msgf("building OpenAI chat completion request v2 message [%d / %d]", i+1, len(v))
			messages = append(messages, msg)
		}
	}

	return messages
}
