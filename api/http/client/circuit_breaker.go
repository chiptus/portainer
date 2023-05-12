package client

import (
	"io"
	"net/http"
	"time"
)

// circuitBreakerURL is the URL of the circuit breaker service that is used to check if the OpenAI integration is disabled.
const circuitBreakerURL = "http://cb-openai.portainer.io/check"

// circuitBreakerTimeout is the timeout used when checking the circuit breaker service.
const circuitBreakerTimeout = 3 * time.Second

// CheckCircuitBreakerForOpenAIIntegration checks if the OpenAI integration is disabled.
// It will return true if the integration is enabled, false otherwise.
func CheckCircuitBreakerForOpenAIIntegration() bool {
	httpcli := &http.Client{Timeout: circuitBreakerTimeout}

	resp, err := httpcli.Head(circuitBreakerURL)

	// The circuit breaker service is either not running anymore or it is answering something else than 204.
	// In both cases, assume that the feature has been remotely disabled.
	if err != nil {
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return resp.StatusCode == http.StatusNoContent
}
