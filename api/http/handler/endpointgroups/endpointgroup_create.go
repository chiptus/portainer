package endpointgroups

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type endpointGroupCreatePayload struct {
	// Environment(Endpoint) group name
	Name string `validate:"required" example:"my-Environment-group"`
	// Environment(Endpoint) group description
	Description string `example:"description"`
	// List of environment(endpoint) identifiers that will be part of this group
	AssociatedEndpoints []portaineree.EndpointID `example:"1,3"`
	// List of tag identifiers to which this environment(endpoint) group is associated
	TagIDs []portaineree.TagID `example:"1,2"`
}

func (payload *endpointGroupCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid environment group name")
	}
	if payload.TagIDs == nil {
		payload.TagIDs = []portaineree.TagID{}
	}
	return nil
}

// @summary Create an Environment(Endpoint) Group
// @description Create a new environment(endpoint) group.
// @description **Access policy**: administrator
// @tags endpoint_groups
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body endpointGroupCreatePayload true "Environment(Endpoint) Group details"
// @success 200 {object} portaineree.EndpointGroup "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /endpoint_groups [post]
func (handler *Handler) endpointGroupCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload endpointGroupCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	endpointGroup := &portaineree.EndpointGroup{
		Name:               payload.Name,
		Description:        payload.Description,
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             payload.TagIDs,
	}

	err = handler.DataStore.EndpointGroup().Create(endpointGroup)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the environment group inside the database", err)
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments from the database", err)
	}

	for _, id := range payload.AssociatedEndpoints {
		for _, endpoint := range endpoints {
			if endpoint.ID == id {
				endpoint.GroupID = endpointGroup.ID

				err := handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, &endpoint)
				if err != nil {
					return httperror.InternalServerError("Unable to update environment", err)
				}

				err = handler.updateEndpointRelations(&endpoint, endpointGroup)
				if err != nil {
					return httperror.InternalServerError("Unable to persist environment relations changes inside the database", err)
				}

				break
			}
		}
	}

	for _, tagID := range endpointGroup.TagIDs {
		tag, err := handler.DataStore.Tag().Tag(tagID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve tag from the database", err)
		}

		tag.EndpointGroups[endpointGroup.ID] = true

		err = handler.DataStore.Tag().UpdateTag(tagID, tag)
		if err != nil {
			return httperror.InternalServerError("Unable to persist tag changes inside the database", err)
		}
	}

	return response.JSON(w, endpointGroup)
}
