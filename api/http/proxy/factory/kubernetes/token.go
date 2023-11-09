package kubernetes

import (
	"fmt"
	"os"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/authorization"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
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

func (manager *tokenManager) SetupUserServiceAccounts(userID int, endpoint *portaineree.Endpoint) error {
	user, err := manager.dataStore.User().Read(portainer.UserID(userID))
	if err != nil || user == nil {
		return err
	}

	endpointRole, err := manager.authService.GetUserEndpointRoleTx(manager.dataStore, userID, int(endpoint.ID))
	if err != nil || endpointRole == nil {
		return err
	}

	namespaces, err := manager.kubecli.GetNamespaces()
	if err != nil {
		return err
	}

	accessPolicies, err := manager.kubecli.GetNamespaceAccessPolicies()
	if err != nil {
		return err
	}
	// update the namespace access policies based on user's role, also in configmap.
	accessPolicies, hasChange, _ := manager.authService.UpdateUserNamespaceAccessPolicies(
		manager.dataStore, userID, endpoint, accessPolicies,
	)
	if hasChange {
		err = manager.kubecli.UpdateNamespaceAccessPolicies(accessPolicies)
		if err != nil {
			return err
		}
	}

	namespaceRoles, err := manager.authService.GetUserNamespaceRoles(
		manager.dataStore,
		userID, int(endpointRole.ID), int(endpoint.ID), accessPolicies, namespaces,
		endpointRole.Authorizations, endpoint.Kubernetes.Configuration,
	)
	if err != nil {
		return err
	}

	err = manager.kubecli.SetupUserServiceAccount(
		*user, endpointRole.ID, namespaces, namespaceRoles, endpoint.Kubernetes.Configuration,
	)
	if err != nil {
		return err
	}

	return nil
}

func (manager *tokenManager) UpdateUserServiceAccountsForEndpoint(endpointID portainer.EndpointID) {
	endpoint, err := manager.dataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		log.Error().Err(err).Msgf("failed fetching environments %d", endpointID)
		return
	}

	userIDs := make([]portainer.UserID, 0)
	for u := range endpoint.UserAccessPolicies {
		userIDs = append(userIDs, u)
	}
	for t := range endpoint.TeamAccessPolicies {
		memberships, _ := manager.dataStore.TeamMembership().TeamMembershipsByTeamID(portainer.TeamID(t))
		for _, membership := range memberships {
			userIDs = append(userIDs, membership.UserID)
		}
	}

	for _, userID := range userIDs {
		err = manager.SetupUserServiceAccounts(int(userID), endpoint)
		if err != nil {
			log.Error().Err(err).Msgf("failed updating user service account for environment %d", endpoint.ID)
		}
	}
}

// GetUserServiceAccountToken setup a user's service account if does not exist, then retrieve its token
func (manager *tokenManager) GetUserServiceAccountToken(
	userID int, endpointID int,
) (string, error) {
	tokenFunc := func() (string, error) {
		endpoint, err := manager.dataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
		if err != nil {
			return "", err
		}

		err = manager.SetupUserServiceAccounts(userID, endpoint)
		if err != nil {
			return "", fmt.Errorf("while updating user service accounts for environment: %w", err)
		}
		return manager.kubecli.GetServiceAccountBearerToken(userID)
	}

	return manager.tokenCache.getOrAddToken(portainer.UserID(userID), tokenFunc)
}
