package kubernetes

import (
	"strconv"
	"sync"

	cmap "github.com/orcaman/concurrent-map"
)

type (
	// TokenCacheManager represents a service used to manage multiple tokenCache objects.
	TokenCacheManager struct {
		tokenCaches cmap.ConcurrentMap
	}

	tokenCache struct {
		userTokenCache cmap.ConcurrentMap
		mutex       sync.Mutex
	}
)

// NewTokenCacheManager returns a pointer to a new instance of TokenCacheManager
func NewTokenCacheManager() *TokenCacheManager {
	return &TokenCacheManager{
		tokenCaches: cmap.New(),
	}
}

// CreateTokenCache will create a new tokenCache object, associate it to the manager map of caches
// and return a pointer to that tokenCache instance.
func (manager *TokenCacheManager) CreateTokenCache(endpointID int) *tokenCache {
	tokenCache := newTokenCache()

	key := strconv.Itoa(endpointID)
	manager.tokenCaches.Set(key, tokenCache)

	return tokenCache
}

// GetOrCreateTokenCache will get the tokenCache from the manager map of caches if it exists,
// otherwise it will create a new tokenCache object, associate it to the manager map of caches
// and return a pointer to that tokenCache instance.
func (manager *TokenCacheManager) GetOrCreateTokenCache(endpointID int) *tokenCache {
	key := strconv.Itoa(endpointID)
	if epCache, ok := manager.tokenCaches.Get(key); ok {
		return epCache.(*tokenCache)
	}

	return manager.CreateTokenCache(endpointID)
}

// RemoveUserFromCache will ensure that the specific userID is removed from all registered caches.
func (manager *TokenCacheManager) RemoveUserFromCache(userID int) {
	for cache := range manager.tokenCaches.IterBuffered() {
		cache.Val.(*tokenCache).removeToken(userID)
	}
}

// remove all tokens in an endpoint
func (manager *TokenCacheManager) HandleEndpointAuthUpdate(endpointID int) {
	key := strconv.Itoa(endpointID)
	if epCache, ok := manager.tokenCaches.Get(key); ok {
		epCache.(*tokenCache).userTokenCache = cmap.New()
	}
}

// remove all user's token when all users' auth are updated
func (manager *TokenCacheManager) HandleUsersAuthUpdate() {
	for cache := range manager.tokenCaches.IterBuffered() {
		cache.Val.(*tokenCache).userTokenCache = cmap.New()
	}
}

// remove a single user token when his auth is updated
func (manager *TokenCacheManager) HandleUserAuthDelete(userID int) {
	manager.RemoveUserFromCache(userID)
}

func newTokenCache() *tokenCache {
	return &tokenCache{
		userTokenCache: cmap.New(),
		mutex:       sync.Mutex{},
	}
}

func (cache *tokenCache) getToken(userID int) (string, bool) {
	key := strconv.Itoa(userID)
	token, ok := cache.userTokenCache.Get(key)
	if ok {
		return token.(string), true
	}

	return "", false
}

func (cache *tokenCache) addToken(userID int, token string) {
	key := strconv.Itoa(userID)
	cache.userTokenCache.Set(key, token)
}

func (cache *tokenCache) removeToken(userID int) {
	key := strconv.Itoa(userID)
	cache.userTokenCache.Remove(key)
}
