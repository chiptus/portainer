package openai

import (
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

type (
	// OpenAIPromptBuilder is responsible for building the prompt that will be sent to OpenAI.
	OpenAIPromptBuilder struct {
		DataStore       dataservices.DataStore
		SnapshotService portaineree.SnapshotService
	}

	// PromptParameters contains the parameters required to build a prompt.
	PromptParameters struct {
		PromptType    PromptType
		UserMessage   string
		EnvironmentID portaineree.EndpointID
		User          portaineree.User
	}

	// PromptType is the type of prompt that will be built.
	PromptType string

	// SupportedModel is a supported OpenAI model.
	// https://platform.openai.com/docs/models/models
	SupportedModel string
)

const (
	// PROMPT_ENVIRONMENT_AWARE is a prompt type that is used to build a prompt that is aware of an environment.
	PROMPT_ENVIRONMENT_AWARE PromptType = "environment_aware"
)

const (
	// MODEL_GPT_4
	MODEL_GPT_4 SupportedModel = openai.GPT4
	// MODEL_GPT_3_5_TURBO
	MODEL_GPT_3_5_TURBO SupportedModel = openai.GPT3Dot5Turbo
	// MODEL_TEXT_DAVINCI_003
	MODEL_TEXT_DAVINCI_003 SupportedModel = openai.GPT3TextDavinci003
)

// SupportedModelList returns a list of supported models as strings.
func SupportedModelList() []string {
	return []string{
		string(MODEL_GPT_4),
		string(MODEL_GPT_3_5_TURBO),
		string(MODEL_TEXT_DAVINCI_003),
	}
}

// NewPromptBuilder creates a new instance of OpenAIPromptBuilder.
func NewPromptBuilder(dataStore dataservices.DataStore, snapshotService portaineree.SnapshotService) OpenAIPromptBuilder {
	return OpenAIPromptBuilder{
		DataStore:       dataStore,
		SnapshotService: snapshotService,
	}
}

// BuildTextCompletionPrompt builds a prompt based on the provided parameters.
// This prompt is designed to work with OpenAI completion requests: https://platform.openai.com/docs/api-reference/completions
// It uses the following prompt model:
//
// base (prefix) prompt
// server aware context
// user aware context
// environment aware context
// ###
// sanitized message
// ###
//
// It will return the prompt as a single string.
func (builder *OpenAIPromptBuilder) BuildTextCompletionPrompt(params PromptParameters) (string, error) {
	prompt := []string{"You must help me use Portainer EE."}

	prompt = append(prompt, buildServerAwareContext())

	prompt = append(prompt, builder.buildUserAwareContext(params.User))

	switch params.PromptType {
	case PROMPT_ENVIRONMENT_AWARE:
		environmentAwareContext, err := builder.buildEnvironmentAwareContext(params.EnvironmentID)
		if err != nil {
			log.Err(err).Msg("unable to create environment aware context when building OpenAI prompt")
			return "", err
		}

		prompt = append(prompt, environmentAwareContext)
	}

	prompt = append(prompt, "###")

	prompt = append(prompt, sanitizeMessage(params.UserMessage))

	prompt = append(prompt, "###")

	return strings.Join(prompt, " "), nil
}

// BuildChatCompletionPrompt builds a prompt based on the provided parameters.
// This prompt is designed to work with OpenAI chat completion requests: https://platform.openai.com/docs/api-reference/chat
// It uses the following prompt model:
//
// system: base (prefix) prompt
// user: server aware context
// user: user aware context
// user: environment aware context
// user: sanitized message
//
// It will return the prompt as a map associating different information sentences (including the user message) to roles.
func (builder *OpenAIPromptBuilder) BuildChatCompletionPrompt(params PromptParameters) (map[string][]string, error) {
	prompt := map[string][]string{
		"system": {"You are a helpful assistant that helps users use Portainer EE to deploy and troubleshoot containerized applications."},
		"user": {
			buildServerAwareContext(),
			builder.buildUserAwareContext(params.User),
		},
	}

	switch params.PromptType {
	case PROMPT_ENVIRONMENT_AWARE:
		environmentAwareContext, err := builder.buildEnvironmentAwareContext(params.EnvironmentID)
		if err != nil {
			log.Err(err).Msg("unable to create environment aware context when building OpenAI prompt")
			return nil, err
		}

		prompt["user"] = append(prompt["user"], environmentAwareContext)
	}

	prompt["user"] = append(prompt["user"], sanitizeMessage(params.UserMessage))
	return prompt, nil
}
