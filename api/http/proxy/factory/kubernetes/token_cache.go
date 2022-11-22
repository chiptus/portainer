package kubernetes

import (
	"sync"

	portaineree "github.com/portainer/portainer-ee/api"
)

// TokenCacheManager represents a service used to manage multiple tokenCache objects.
type TokenCacheManager struct {
	tokenCaches map[portaineree.EndpointID]*tokenCache
	mu          sync.Mutex
}

type tokenCache struct {
	userTokenCache map[portaineree.UserID]string
	mu             sync.Mutex
}

// NewTokenCacheManager returns a pointer to a new instance of TokenCacheManager
func NewTokenCacheManager() *TokenCacheManager {
	return &TokenCacheManager{
		tokenCaches: make(map[portaineree.EndpointID]*tokenCache),
	}
}

// GetOrCreateTokenCache will get the tokenCache from the manager map of caches if it exists,
// otherwise it will create a new tokenCache object, associate it to the manager map of caches
// and return a pointer to that tokenCache instance.
func (manager *TokenCacheManager) GetOrCreateTokenCache(endpointID portaineree.EndpointID) *tokenCache {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if tc, ok := manager.tokenCaches[endpointID]; ok {
		return tc
	}

	tc := &tokenCache{
		userTokenCache: make(map[portaineree.UserID]string),
	}

	manager.tokenCaches[endpointID] = tc

	return tc
}

// RemoveUserFromCache will ensure that the specific userID is removed from all registered caches.
func (manager *TokenCacheManager) RemoveUserFromCache(userID portaineree.UserID) {
	manager.mu.Lock()
	for _, tc := range manager.tokenCaches {
		tc.removeToken(userID)
	}
	manager.mu.Unlock()
}

// HandleEndpointAuthUpdate remove all tokens in an endpoint
func (manager *TokenCacheManager) HandleEndpointAuthUpdate(endpointID portaineree.EndpointID) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if _, ok := manager.tokenCaches[endpointID]; !ok {
		return
	}

	manager.tokenCaches[endpointID].userTokenCache = make(map[portaineree.UserID]string)
}

// HandleUsersAuthUpdate remove all user's token when all users' auth are updated
func (manager *TokenCacheManager) HandleUsersAuthUpdate() {
	manager.mu.Lock()
	for _, cache := range manager.tokenCaches {
		cache.userTokenCache = make(map[portaineree.UserID]string)
	}
	manager.mu.Unlock()
}

// remove a single user token when his auth is updated
func (manager *TokenCacheManager) HandleUserAuthDelete(userID portaineree.UserID) {
	manager.RemoveUserFromCache(userID)
}

func (cache *tokenCache) getOrAddToken(userID portaineree.UserID, tokenGetFunc func() (string, error)) (string, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if tok, ok := cache.userTokenCache[userID]; ok {
		return tok, nil
	}

	tok, err := tokenGetFunc()
	if err != nil {
		return "", err
	}

	cache.userTokenCache[userID] = tok

	return tok, nil
}

func (cache *tokenCache) removeToken(userID portaineree.UserID) {
	cache.mu.Lock()
	delete(cache.userTokenCache, userID)
	cache.mu.Unlock()
}
