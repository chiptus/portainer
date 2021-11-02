package endpoints

import (
	"net/http"

	werrors "github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/http/useractivity"
)

type endpointSettingsUpdatePayload struct {
	// Security settings updates
	SecuritySettings struct {
		// Whether non-administrator should be able to use bind mounts when creating containers
		AllowBindMountsForRegularUsers *bool `json:"allowBindMountsForRegularUsers" example:"false"`
		// Whether non-administrator should be able to use privileged mode when creating containers
		AllowPrivilegedModeForRegularUsers *bool `json:"allowPrivilegedModeForRegularUsers" example:"false"`
		// Whether non-administrator should be able to browse volumes
		AllowVolumeBrowserForRegularUsers *bool `json:"allowVolumeBrowserForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use the host pid
		AllowHostNamespaceForRegularUsers *bool `json:"allowHostNamespaceForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use device mapping
		AllowDeviceMappingForRegularUsers *bool `json:"allowDeviceMappingForRegularUsers" example:"true"`
		// Whether non-administrator should be able to manage stacks
		AllowStackManagementForRegularUsers *bool `json:"allowStackManagementForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use container capabilities
		AllowContainerCapabilitiesForRegularUsers *bool `json:"allowContainerCapabilitiesForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use sysctl settings
		AllowSysctlSettingForRegularUsers *bool `json:"allowSysctlSettingForRegularUsers" example:"true"`
		// Whether host management features are enabled
		EnableHostManagementFeatures *bool `json:"enableHostManagementFeatures" example:"true"`
	} `json:"securitySettings"`
	// Whether automatic update time restrictions are enabled
	ChangeWindow *portainer.EndpointChangeWindow `json:"changeWindow"`
}

func (payload *endpointSettingsUpdatePayload) Validate(_ *http.Request) error {
	if payload.ChangeWindow != nil {
		err := validateAutoUpdateSettings(*payload.ChangeWindow)
		if err != nil {
			return werrors.WithMessage(err, "Validation failed")
		}
	}

	return nil
}

// @id EndpointSettingsUpdate
// @summary Update settings for an environment(endpoint)
// @description Update settings for an environment(endpoint).
// @description **Access policy**: authenticated
// @security jwt
// @tags endpoints
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param body body endpointSettingsUpdatePayload true "Environment(Endpoint) details"
// @success 200 {object} portainer.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/settings [put]
func (handler *Handler) endpointSettingsUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment identifier route variable", err}
	}

	var payload endpointSettingsUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == errors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access environment", err}
	}

	securitySettings := endpoint.SecuritySettings

	if payload.SecuritySettings.AllowBindMountsForRegularUsers != nil {
		securitySettings.AllowBindMountsForRegularUsers = *payload.SecuritySettings.AllowBindMountsForRegularUsers
	}

	if payload.SecuritySettings.AllowContainerCapabilitiesForRegularUsers != nil {
		securitySettings.AllowContainerCapabilitiesForRegularUsers = *payload.SecuritySettings.AllowContainerCapabilitiesForRegularUsers
	}

	if payload.SecuritySettings.AllowDeviceMappingForRegularUsers != nil {
		securitySettings.AllowDeviceMappingForRegularUsers = *payload.SecuritySettings.AllowDeviceMappingForRegularUsers
	}

	if payload.SecuritySettings.AllowHostNamespaceForRegularUsers != nil {
		securitySettings.AllowHostNamespaceForRegularUsers = *payload.SecuritySettings.AllowHostNamespaceForRegularUsers
	}

	if payload.SecuritySettings.AllowPrivilegedModeForRegularUsers != nil {
		securitySettings.AllowPrivilegedModeForRegularUsers = *payload.SecuritySettings.AllowPrivilegedModeForRegularUsers
	}

	if payload.SecuritySettings.AllowStackManagementForRegularUsers != nil {
		securitySettings.AllowStackManagementForRegularUsers = *payload.SecuritySettings.AllowStackManagementForRegularUsers
	}

	updateAuthorizations := false
	if payload.SecuritySettings.AllowVolumeBrowserForRegularUsers != nil {
		updateAuthorizations = securitySettings.AllowVolumeBrowserForRegularUsers != *payload.SecuritySettings.AllowVolumeBrowserForRegularUsers
		securitySettings.AllowVolumeBrowserForRegularUsers = *payload.SecuritySettings.AllowVolumeBrowserForRegularUsers
	}

	if payload.SecuritySettings.AllowSysctlSettingForRegularUsers != nil {
		securitySettings.AllowSysctlSettingForRegularUsers = *payload.SecuritySettings.AllowSysctlSettingForRegularUsers
	}

	if payload.SecuritySettings.EnableHostManagementFeatures != nil {
		securitySettings.EnableHostManagementFeatures = *payload.SecuritySettings.EnableHostManagementFeatures
	}

	endpoint.SecuritySettings = securitySettings

	if payload.ChangeWindow != nil {
		endpoint.ChangeWindow = *payload.ChangeWindow
	}

	err = handler.DataStore.Endpoint().UpdateEndpoint(portainer.EndpointID(endpointID), endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Failed persisting environment in database", err}
	}

	if payload.ChangeWindow != nil {
		// Make it clear that the time stored in the user activity log is actually UTC despite, not the timezone value
		payload.ChangeWindow.StartTime = payload.ChangeWindow.StartTime + " [UTC]"
		payload.ChangeWindow.EndTime = payload.ChangeWindow.EndTime + " [UTC]"
	}

	useractivity.LogHttpActivity(handler.UserActivityStore, endpoint.Name, r, payload)

	if updateAuthorizations {
		err := handler.AuthorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update RBAC authorizations", err}
		}
	}

	return response.JSON(w, endpoint)
}
