package endpointgroups

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer/pkg/featureflags"
)

// @id EndpointGroupDelete
// @summary Remove an environment(endpoint) group
// @description Remove an environment(endpoint) group.
// @description **Access policy**: administrator
// @tags endpoint_groups
// @security ApiKeyAuth
// @security jwt
// @param id path int true "EndpointGroup identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 404 "EndpointGroup not found"
// @failure 500 "Server error"
// @router /endpoint_groups/{id} [delete]
func (handler *Handler) endpointGroupDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment group identifier route variable", err)
	}

	if endpointGroupID == 1 {
		return httperror.Forbidden("Unable to remove the default 'Unassigned' group", errors.New("cannot remove the default environment group"))
	}

	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		err = handler.deleteEndpointGroup(handler.DataStore, portaineree.EndpointGroupID(endpointGroupID))
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			return handler.deleteEndpointGroup(tx, portaineree.EndpointGroupID(endpointGroupID))
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.Empty(w)
}

func (handler *Handler) deleteEndpointGroup(tx dataservices.DataStoreTx, endpointGroupID portaineree.EndpointGroupID) error {
	endpointGroup, err := tx.EndpointGroup().EndpointGroup(endpointGroupID)
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment group with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment group with the specified identifier inside the database", err)
	}

	err = tx.EndpointGroup().DeleteEndpointGroup(endpointGroupID)
	if err != nil {
		return httperror.InternalServerError("Unable to remove the environment group from the database", err)
	}

	endpoints, err := tx.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments from the database", err)
	}

	updateAuthorizations := false
	for _, endpoint := range endpoints {
		if endpoint.GroupID == endpointGroupID {
			updateAuthorizations = true
			endpoint.GroupID = 1
			err = tx.Endpoint().UpdateEndpoint(endpoint.ID, &endpoint)
			if err != nil {
				return httperror.InternalServerError("Unable to update environment", err)
			}

			err = handler.updateEndpointRelations(tx, &endpoint, nil)
			if err != nil {
				return httperror.InternalServerError("Unable to persist environment relations changes inside the database", err)
			}
		}
	}

	if updateAuthorizations {
		err = handler.AuthorizationService.UpdateUsersAuthorizationsTx(tx)
		if err != nil {
			return httperror.InternalServerError("Unable to update user authorizations", err)
		}
	}

	for _, tagID := range endpointGroup.TagIDs {
		if featureflags.IsEnabled(portaineree.FeatureNoTx) {
			err = tx.Tag().UpdateTagFunc(tagID, func(tag *portaineree.Tag) {
				delete(tag.EndpointGroups, endpointGroup.ID)
			})

			if tx.IsErrObjectNotFound(err) {
				return httperror.InternalServerError("Unable to find a tag inside the database", err)
			} else if err != nil {
				return httperror.InternalServerError("Unable to persist tag changes inside the database", err)
			}

			continue
		}

		tag, err := tx.Tag().Tag(tagID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve tag from the database", err)
		}

		delete(tag.EndpointGroups, endpointGroup.ID)

		err = tx.Tag().UpdateTag(tagID, tag)
		if err != nil {
			return httperror.InternalServerError("Unable to persist tag changes inside the database", err)
		}
	}

	return nil
}
