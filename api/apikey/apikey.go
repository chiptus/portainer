package apikey

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

// APIKeyService represents a service for managing API keys.
type APIKeyService interface {
	HashRaw(rawKey string) []byte
	GenerateApiKey(user portaineree.User, description string) (string, *portaineree.APIKey, error)
	GetAPIKey(apiKeyID portainer.APIKeyID) (*portaineree.APIKey, error)
	GetAPIKeys(userID portainer.UserID) ([]portaineree.APIKey, error)
	GetDigestUserAndKey(digest []byte) (portaineree.User, portaineree.APIKey, error)
	UpdateAPIKey(apiKey *portaineree.APIKey) error
	DeleteAPIKey(apiKeyID portainer.APIKeyID) error
	InvalidateUserKeyCache(userId portainer.UserID) bool
}
