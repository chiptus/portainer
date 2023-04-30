package openai

import "strings"

// sanitizeMessage sanitizes the message that will be sent to the OpenAI API.
func sanitizeMessage(message string) string {
	if !strings.HasSuffix(message, ".") {
		message += "."
	}

	return message
}
