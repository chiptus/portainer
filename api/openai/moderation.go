package openai

import (
	"context"
	"time"

	measure "github.com/portainer/portainer-ee/api/internal/time"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

// SendModerationRequest sends a moderation request to the OpenAI API.
// https://platform.openai.com/docs/guides/moderation/overview
func SendModerationRequest(context context.Context, userMessage, apiKey string) (bool, error) {
	client := openai.NewClient(apiKey)

	log.Debug().Str("Message", userMessage).Msg("sending OpenAI moderation request")
	defer measure.TrackTime(time.Now(), "OpenAI moderation request")

	moderation, err := client.Moderations(context, openai.ModerationRequest{
		Input: userMessage,
	})
	if err != nil {
		log.Error().Err(err).Msg("an error occured while sending OpenAI moderation request")
		return false, err
	}

	for _, modResult := range moderation.Results {
		if modResult.Flagged {
			log.Error().Str("Message", userMessage).Interface("Categories", modResult.Categories).Msg("OpenAI moderation flagged prompt")
			return true, nil
		}
	}

	return false, nil
}
