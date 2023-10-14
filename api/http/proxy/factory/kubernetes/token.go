package kubernetes

import (
	"os"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"
)

const defaultServiceAccountTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

type tokenManager struct {
	tokenCache  *tokenCache
	kubecli     portaineree.KubeClient
	dataStore   dataservices.DataStore
	adminToken  string
	authService *authorization.Service
}

// NewTokenManager returns a pointer to a new instance of tokenManager.
// If the useLocalAdminToken parameter is set to true, it will search for the local admin service account
// and associate it to the manager.
func NewTokenManager(
	kubecli portaineree.KubeClient,
	dataStore dataservices.DataStore,
	cache *tokenCache,
	setLocalAdminToken bool,
	authService *authorization.Service,
) (*tokenManager, error) {
	tokenManager := &tokenManager{
		tokenCache:  cache,
		kubecli:     kubecli,
		dataStore:   dataStore,
		adminToken:  "",
		authService: authService,
	}

	if setLocalAdminToken {
		token, err := os.ReadFile(defaultServiceAccountTokenFile)
		if err != nil {
			return nil, err
		}

		tokenManager.adminToken = string(token)
	}

	return tokenManager, nil
}

func (manager *tokenManager) GetAdminServiceAccountToken() string {
	return manager.adminToken
}

// GetUserServiceAccountToken setup a user's service account if does not exist, then retrieve its token
func (manager *tokenManager) GetUserServiceAccountToken(
	userID int, endpointID int,
) (string, error) {
	tokenFunc := func() (string, error) {
		user, err := manager.dataStore.User().Read(portainer.UserID(userID))
		if err != nil || user == nil {
			return "", err
		}

		endpointRole, err := manager.authService.GetUserEndpointRoleTx(manager.dataStore, userID, endpointID)
		if err != nil || endpointRole == nil {
			return "", err
		}

		endpoint, err := manager.dataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
		if err != nil {
			return "", err
		}

		namespaces, err := manager.kubecli.GetNamespaces()
		if err != nil {
			return "", err
		}

		accessPolicies, err := manager.kubecli.GetNamespaceAccessPolicies()
		if err != nil {
			return "", err
		}
		// update the namespace access policies based on user's role, also in configmap.
		accessPolicies, hasChange, err := manager.authService.UpdateUserNamespaceAccessPolicies(
			manager.dataStore, userID, endpoint, accessPolicies,
		)
		if hasChange {
			err = manager.kubecli.UpdateNamespaceAccessPolicies(accessPolicies)
			if err != nil {
				return "", err
			}
		}

		namespaceRoles, err := manager.authService.GetUserNamespaceRoles(
			manager.dataStore,
			userID, int(endpointRole.ID), endpointID, accessPolicies, namespaces,
			endpointRole.Authorizations, endpoint.Kubernetes.Configuration,
		)
		if err != nil {
			return "", err
		}

		err = manager.kubecli.SetupUserServiceAccount(
			*user, endpointRole.ID, namespaces, namespaceRoles, endpoint.Kubernetes.Configuration,
		)
		if err != nil {
			return "", err
		}

		return manager.kubecli.GetServiceAccountBearerToken(userID)
	}

	return manager.tokenCache.getOrAddToken(portainer.UserID(userID), tokenFunc)
}
