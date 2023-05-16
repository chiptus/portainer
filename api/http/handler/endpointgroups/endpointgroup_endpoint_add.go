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

// @id EndpointGroupAddEndpoint
// @summary Add an environment(endpoint) to an environment(endpoint) group
// @description Add an environment(endpoint) to an environment(endpoint) group
// @description **Access policy**: administrator
// @tags endpoint_groups
// @security ApiKeyAuth
// @security jwt
// @param id path int true "EndpointGroup identifier"
// @param endpointId path int true "Environment(Endpoint) identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 404 "EndpointGroup not found"
// @failure 500 "Server error"
// @router /endpoint_groups/{id}/endpoints/{endpointId} [put]
func (handler *Handler) endpointGroupAddEndpoint(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment group identifier route variable", err)
	}

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "endpointId")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		err = handler.addEndpoint(handler.DataStore, portaineree.EndpointGroupID(endpointGroupID), portaineree.EndpointID(endpointID))
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			return handler.addEndpoint(tx, portaineree.EndpointGroupID(endpointGroupID), portaineree.EndpointID(endpointID))
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

func (handler *Handler) addEndpoint(tx dataservices.DataStoreTx, endpointGroupID portaineree.EndpointGroupID, endpointID portaineree.EndpointID) error {
	endpointGroup, err := tx.EndpointGroup().EndpointGroup(endpointGroupID)
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment group with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment group with the specified identifier inside the database", err)
	}

	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	endpoint.GroupID = endpointGroup.ID

	err = tx.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
	}

	err = handler.updateEndpointRelations(tx, endpoint, endpointGroup)
	if err != nil {
		return httperror.InternalServerError("Unable to persist environment relations changes inside the database", err)
	}

	err = handler.AuthorizationService.UpdateUsersAuthorizationsTx(tx)
	if err != nil {
		return httperror.InternalServerError("Unable to update user authorizations", err)
	}

	return nil
}
