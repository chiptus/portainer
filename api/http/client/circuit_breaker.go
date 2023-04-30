package client

import (
	"net/http"
	"time"
)

// CIRCUIT_BREAKER_URL is the URL of the circuit breaker service that is used to check if the OpenAI integration is disabled.
const CIRCUIT_BREAKER_URL = "http://cb-openai.portainer.io/check"

// CIRCUIT_BREAKER_TIMEOUT is the timeout used when checking the circuit breaker service.
const CIRCUIT_BREAKER_TIMEOUT = 3 * time.Second

// CheckCircuitBreakerForOpenAIIntegration checks if the OpenAI integration is disabled.
// It will return true if the integration is enabled, false otherwise.
func CheckCircuitBreakerForOpenAIIntegration() bool {
	httpcli := http.Client{Timeout: CIRCUIT_BREAKER_TIMEOUT}

	cbRes, err := httpcli.Head(CIRCUIT_BREAKER_URL)
	if err != nil || (err == nil && cbRes.StatusCode != http.StatusNoContent) {
		// The circuit breaker service is either not running anymore or it is answering something else than 204.
		// In both cases, assume that the feature has been remotely disabled.
		return false
	}

	return true
}
