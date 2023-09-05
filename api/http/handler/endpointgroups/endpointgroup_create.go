package endpointgroups

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer/pkg/featureflags"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/asaskevich/govalidator"
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
		return errors.New("invalid environment group name")
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

	var endpointGroup *portaineree.EndpointGroup
	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		endpointGroup, err = handler.createEndpointGroup(handler.DataStore, payload)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			endpointGroup, err = handler.createEndpointGroup(tx, payload)
			return err
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, endpointGroup)
}

func (handler *Handler) createEndpointGroup(tx dataservices.DataStoreTx, payload endpointGroupCreatePayload) (*portaineree.EndpointGroup, error) {
	endpointGroup := &portaineree.EndpointGroup{
		Name:               payload.Name,
		Description:        payload.Description,
		UserAccessPolicies: portaineree.UserAccessPolicies{},
		TeamAccessPolicies: portaineree.TeamAccessPolicies{},
		TagIDs:             payload.TagIDs,
	}

	err := tx.EndpointGroup().Create(endpointGroup)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to persist the environment group inside the database", err)
	}

	endpoints, err := tx.Endpoint().Endpoints()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve environments from the database", err)
	}

	for _, id := range payload.AssociatedEndpoints {
		for _, endpoint := range endpoints {
			if endpoint.ID == id {
				endpoint.GroupID = endpointGroup.ID

				err := tx.Endpoint().UpdateEndpoint(endpoint.ID, &endpoint)
				if err != nil {
					return nil, httperror.InternalServerError("Unable to update environment", err)
				}

				err = handler.updateEndpointRelations(tx, &endpoint, endpointGroup)
				if err != nil {
					return nil, httperror.InternalServerError("Unable to persist environment relations changes inside the database", err)
				}

				break
			}
		}
	}

	for _, tagID := range endpointGroup.TagIDs {
		if featureflags.IsEnabled(portaineree.FeatureNoTx) {
			err = tx.Tag().UpdateTagFunc(tagID, func(tag *portaineree.Tag) {
				tag.EndpointGroups[endpointGroup.ID] = true
			})

			if tx.IsErrObjectNotFound(err) {
				return nil, httperror.InternalServerError("Unable to find a tag inside the database", err)
			} else if err != nil {
				return nil, httperror.InternalServerError("Unable to persist tag changes inside the database", err)
			}

			continue
		}

		tag, err := tx.Tag().Read(tagID)
		if tx.IsErrObjectNotFound(err) {
			return nil, httperror.InternalServerError("Unable to find a tag inside the database", err)
		}

		tag.EndpointGroups[endpointGroup.ID] = true

		err = tx.Tag().Update(tag.ID, tag)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to persist tag changes inside the database", err)
		}
	}

	return endpointGroup, nil
}
