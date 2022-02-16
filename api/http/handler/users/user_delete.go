package users

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

// @id UserDelete
// @summary Remove a user
// @description Remove a user.
// @description **Access policy**: administrator
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id} [delete]
func (handler *Handler) userDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid user identifier route variable", err}
	}

	if userID == 1 {
		return &httperror.HandlerError{http.StatusForbidden, "Cannot remove the initial admin account", errors.New("Cannot remove the initial admin account")}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user authentication token", err}
	}

	if tokenData.ID == portaineree.UserID(userID) {
		return &httperror.HandlerError{http.StatusForbidden, "Cannot remove your own user account. Contact another administrator", errAdminCannotRemoveSelf}
	}

	user, err := handler.DataStore.User().User(portaineree.UserID(userID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a user with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a user with the specified identifier inside the database", err}
	}

	if user.Role == portaineree.AdministratorRole {
		return handler.deleteAdminUser(w, user)
	}

	handler.AuthorizationService.TriggerUserAuthUpdate(int(user.ID))

	return handler.deleteUser(w, user)
}

func (handler *Handler) deleteAdminUser(w http.ResponseWriter, user *portaineree.User) *httperror.HandlerError {
	if user.Password == "" {
		return handler.deleteUser(w, user)
	}

	users, err := handler.DataStore.User().Users()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve users from the database", err}
	}

	localAdminCount := 0
	for _, u := range users {
		if u.Role == portaineree.AdministratorRole && u.Password != "" {
			localAdminCount++
		}
	}

	if localAdminCount < 2 {
		return &httperror.HandlerError{http.StatusInternalServerError, "Cannot remove local administrator user", errCannotRemoveLastLocalAdmin}
	}

	return handler.deleteUser(w, user)
}

func (handler *Handler) deleteUser(w http.ResponseWriter, user *portaineree.User) *httperror.HandlerError {
	err := handler.removeUserKubeResources(user)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove user k8s resources", err}
	}

	err = handler.AuthorizationService.RemoveUserAccessPolicies(user.ID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to clean-up user access policies", err}
	}

	err = handler.DataStore.User().DeleteUser(user.ID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove user from the database", err}
	}

	err = handler.DataStore.TeamMembership().DeleteTeamMembershipByUserID(user.ID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove user memberships from the database", err}
	}

	// Remove all of the users persisted API keys
	apiKeys, err := handler.apiKeyService.GetAPIKeys(user.ID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user API keys from the database", err}
	}
	for _, k := range apiKeys {
		err = handler.apiKeyService.DeleteAPIKey(k.ID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove user API key from the database", err}
		}
	}

	return response.Empty(w)
}

func (handler *Handler) removeUserKubeResources(user *portaineree.User) error {
	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return err
	}

	errs := []string{}
	// removes user's k8s service account and all related resources
	for _, endpoint := range endpoints {
		if endpoint.Type != portaineree.KubernetesLocalEnvironment &&
			endpoint.Type != portaineree.AgentOnKubernetesEnvironment &&
			endpoint.Type != portaineree.EdgeAgentOnKubernetesEnvironment {
			continue
		}

		kcl, err := handler.K8sClientFactory.GetKubeClient(&endpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("Unable to get k8s environment access @ %d: %w", int(endpoint.ID), err).Error())
			continue
		}

		err = kcl.RemoveUserServiceAccount(int(user.ID))
		if err != nil {
			errs = append(errs, fmt.Errorf("Unable to remove user service account @ %d: %w", int(endpoint.ID), err).Error())
		}

		accessPolicies, err := kcl.GetNamespaceAccessPolicies()
		if err != nil {
			errs = append(errs, fmt.Errorf("Unable to get environment namespace access @ %d: %w", int(endpoint.ID), err).Error())
			continue
		}

		accessPolicies, hasChange, err := handler.AuthorizationService.RemoveUserNamespaceAccessPolicies(
			int(user.ID), int(endpoint.ID), accessPolicies,
		)
		if hasChange {
			err = kcl.UpdateNamespaceAccessPolicies(accessPolicies)
			if err != nil {
				errs = append(errs, fmt.Errorf("Unable to update environment namespace access @ %d: %w", int(endpoint.ID), err).Error())
				continue
			}
		}
	}

	if len(errs) > 0 {
		err = fmt.Errorf(strings.Join(errs, "\n"))
		logrus.WithError(err).Error("failed to remove user k8s resources")
		// ignore error
	}

	return nil
}
