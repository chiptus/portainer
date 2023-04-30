package openai

import (
	"context"
	"time"

	measure "github.com/portainer/portainer-ee/api/internal/time"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

// SendTextCompletionRequest sends a text completion request to the OpenAI API.
// https://platform.openai.com/docs/guides/completion
func SendTextCompletionRequest(context context.Context, model, prompt, apiKey string) (string, error) {
	client := openai.NewClient(apiKey)

	log.Debug().Str("Prompt", prompt).Msg("sending OpenAI text completion request")
	defer measure.TrackTime(time.Now(), "OpenAI text completion request")

	resp, err := client.CreateCompletion(
		context,
		openai.CompletionRequest{
			Model:       model,
			MaxTokens:   2048,
			TopP:        0.25,
			Prompt:      prompt,
			Temperature: 0,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("an error occured while sending OpenAI text completion request")
		return "", err
	}

	return resp.Choices[0].Text, nil
}
