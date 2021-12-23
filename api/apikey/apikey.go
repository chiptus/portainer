package apikey

import (
	"crypto/rand"
	"io"

	portaineree "github.com/portainer/portainer-ee/api"
)

// APIKeyService represents a service for managing API keys.
type APIKeyService interface {
	HashRaw(rawKey string) []byte
	GenerateApiKey(user portaineree.User, description string) (string, *portaineree.APIKey, error)
	GetAPIKey(apiKeyID portaineree.APIKeyID) (*portaineree.APIKey, error)
	GetAPIKeys(userID portaineree.UserID) ([]portaineree.APIKey, error)
	GetDigestUserAndKey(digest []byte) (portaineree.User, portaineree.APIKey, error)
	UpdateAPIKey(apiKey *portaineree.APIKey) error
	DeleteAPIKey(apiKeyID portaineree.APIKeyID) error
	InvalidateUserKeyCache(userId portaineree.UserID) bool
}

// generateRandomKey generates a random key of specified length
// source: https://github.com/gorilla/securecookie/blob/master/securecookie.go#L515
func generateRandomKey(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}
