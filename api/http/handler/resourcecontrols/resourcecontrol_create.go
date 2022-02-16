package resourcecontrols

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type resourceControlCreatePayload struct {
	//
	ResourceID string `example:"617c5f22bb9b023d6daab7cba43a57576f83492867bc767d1c59416b065e5f08" validate:"required"`
	// Type of Docker resource. Valid values are: container, volume\
	// service, secret, config or stack
	Type string `example:"container" validate:"required"`
	// Permit access to the associated resource to any user
	Public bool `example:"true"`
	// Permit access to resource only to admins
	AdministratorsOnly bool `example:"true"`
	// List of user identifiers with access to the associated resource
	Users []int `example:"1,4"`
	// List of team identifiers with access to the associated resource
	Teams []int `example:"56,7"`
	// List of Docker resources that will inherit this access control
	SubResourceIDs []string `example:"617c5f22bb9b023d6daab7cba43a57576f83492867bc767d1c59416b065e5f08"`
}

var (
	errResourceControlAlreadyExists = errors.New("A resource control is already applied on this resource") //http/resourceControl
	errInvalidResourceControlType   = errors.New("Unsupported resource control type")                      //http/resourceControl
)

func (payload *resourceControlCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.ResourceID) {
		return errors.New("invalid payload: invalid resource identifier")
	}

	if govalidator.IsNull(payload.Type) {
		return errors.New("invalid payload: invalid type")
	}

	if len(payload.Users) == 0 && len(payload.Teams) == 0 && !payload.Public && !payload.AdministratorsOnly {
		return errors.New("invalid payload: must specify Users, Teams, Public or AdministratorsOnly")
	}

	if payload.Public && payload.AdministratorsOnly {
		return errors.New("invalid payload: cannot set both public and administrators only flags to true")
	}
	return nil
}

// @id ResourceControlCreate
// @summary Create a new resource control
// @description Create a new resource control to restrict access to a Docker resource.
// @description **Access policy**: administrator
// @tags resource_controls
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body resourceControlCreatePayload true "Resource control details"
// @success 200 {object} portaineree.ResourceControl "Success"
// @failure 400 "Invalid request"
// @failure 409 "Resource control already exists"
// @failure 500 "Server error"
// @router /resource_controls [post]
func (handler *Handler) resourceControlCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload resourceControlCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	var resourceControlType portaineree.ResourceControlType
	switch payload.Type {
	case "container":
		resourceControlType = portaineree.ContainerResourceControl
	case "container-group":
		resourceControlType = portaineree.ContainerGroupResourceControl
	case "service":
		resourceControlType = portaineree.ServiceResourceControl
	case "volume":
		resourceControlType = portaineree.VolumeResourceControl
	case "network":
		resourceControlType = portaineree.NetworkResourceControl
	case "secret":
		resourceControlType = portaineree.SecretResourceControl
	case "stack":
		resourceControlType = portaineree.StackResourceControl
	case "config":
		resourceControlType = portaineree.ConfigResourceControl
	default:
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid type value. Value must be one of: container, service, volume, network, secret, stack or config", errInvalidResourceControlType}
	}

	rc, err := handler.dataStore.ResourceControl().ResourceControlByResourceIDAndType(payload.ResourceID, resourceControlType)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve resource controls from the database", err}
	}
	if rc != nil {
		return &httperror.HandlerError{http.StatusConflict, "A resource control is already associated to this resource", errResourceControlAlreadyExists}
	}

	var userAccesses = make([]portaineree.UserResourceAccess, 0)
	for _, v := range payload.Users {
		userAccess := portaineree.UserResourceAccess{
			UserID:      portaineree.UserID(v),
			AccessLevel: portaineree.ReadWriteAccessLevel,
		}
		userAccesses = append(userAccesses, userAccess)
	}

	var teamAccesses = make([]portaineree.TeamResourceAccess, 0)
	for _, v := range payload.Teams {
		teamAccess := portaineree.TeamResourceAccess{
			TeamID:      portaineree.TeamID(v),
			AccessLevel: portaineree.ReadWriteAccessLevel,
		}
		teamAccesses = append(teamAccesses, teamAccess)
	}

	resourceControl := portaineree.ResourceControl{
		ResourceID:         payload.ResourceID,
		SubResourceIDs:     payload.SubResourceIDs,
		Type:               resourceControlType,
		Public:             payload.Public,
		AdministratorsOnly: payload.AdministratorsOnly,
		UserAccesses:       userAccesses,
		TeamAccesses:       teamAccesses,
	}

	err = handler.dataStore.ResourceControl().Create(&resourceControl)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the resource control inside the database", err}
	}

	return response.JSON(w, resourceControl)
}
