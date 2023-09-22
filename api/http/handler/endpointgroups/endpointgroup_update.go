package endpointgroups

import (
	"errors"
	"net/http"
	"reflect"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/tag"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/rs/zerolog/log"
)

type endpointGroupUpdatePayload struct {
	// Environment(Endpoint) group name
	Name string `example:"my-environment-group"`
	// Environment(Endpoint) group description
	Description string `example:"description"`
	// List of tag identifiers associated to the environment(endpoint) group
	TagIDs             []portaineree.TagID `example:"3,4"`
	UserAccessPolicies portaineree.UserAccessPolicies
	TeamAccessPolicies portaineree.TeamAccessPolicies
}

func (payload *endpointGroupUpdatePayload) Validate(r *http.Request) error {
	return nil
}

// @id EndpointGroupUpdate
// @summary Update an environment(endpoint) group
// @description Update an environment(endpoint) group.
// @description **Access policy**: administrator
// @tags endpoint_groups
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "EndpointGroup identifier"
// @param body body endpointGroupUpdatePayload true "EndpointGroup details"
// @success 200 {object} portaineree.EndpointGroup "Success"
// @failure 400 "Invalid request"
// @failure 404 "EndpointGroup not found"
// @failure 500 "Server error"
// @router /endpoint_groups/{id} [put]
func (handler *Handler) endpointGroupUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment group identifier route variable", err)
	}

	var payload endpointGroupUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var endpointGroup *portaineree.EndpointGroup
	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		endpointGroup, err = handler.updateEndpointGroup(tx, portaineree.EndpointGroupID(endpointGroupID), payload)
		return err
	})
	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, endpointGroup)
}

func (handler *Handler) updateEndpointGroup(tx dataservices.DataStoreTx, endpointGroupID portaineree.EndpointGroupID, payload endpointGroupUpdatePayload) (*portaineree.EndpointGroup, error) {
	endpointGroup, err := tx.EndpointGroup().Read(portaineree.EndpointGroupID(endpointGroupID))
	if tx.IsErrObjectNotFound(err) {
		return nil, httperror.NotFound("Unable to find an environment group with the specified identifier inside the database", err)
	} else if err != nil {
		return nil, httperror.InternalServerError("Unable to find an environment group with the specified identifier inside the database", err)
	}

	if payload.Name != "" {
		endpointGroup.Name = payload.Name
	}

	if payload.Description != "" {
		endpointGroup.Description = payload.Description
	}

	tagsChanged := false
	if payload.TagIDs != nil {
		payloadTagSet := tag.Set(payload.TagIDs)
		endpointGroupTagSet := tag.Set((endpointGroup.TagIDs))
		union := tag.Union(payloadTagSet, endpointGroupTagSet)
		intersection := tag.Intersection(payloadTagSet, endpointGroupTagSet)
		tagsChanged = len(union) > len(intersection)

		if tagsChanged {
			removeTags := tag.Difference(endpointGroupTagSet, payloadTagSet)

			for tagID := range removeTags {
				tag, err := tx.Tag().Read(portaineree.TagID(tagID))
				if tx.IsErrObjectNotFound(err) {
					return nil, httperror.InternalServerError("Unable to find a tag inside the database", err)
				}

				delete(tag.EndpointGroups, endpointGroup.ID)

				err = tx.Tag().Update(tagID, tag)
				if err != nil {
					return nil, httperror.InternalServerError("Unable to persist tag changes inside the database", err)
				}
			}

			endpointGroup.TagIDs = payload.TagIDs
			for _, tagID := range payload.TagIDs {
				tag, err := tx.Tag().Read(portaineree.TagID(tagID))
				if tx.IsErrObjectNotFound(err) {
					return nil, httperror.InternalServerError("Unable to find a tag inside the database", err)
				}

				tag.EndpointGroups[endpointGroup.ID] = true

				err = tx.Tag().Update(tagID, tag)
				if err != nil {
					return nil, httperror.InternalServerError("Unable to persist tag changes inside the database", err)
				}
			}
		}
	}

	updateAuthorizations := false
	if payload.UserAccessPolicies != nil && !reflect.DeepEqual(payload.UserAccessPolicies, endpointGroup.UserAccessPolicies) {
		endpointGroup.UserAccessPolicies = payload.UserAccessPolicies
		updateAuthorizations = true
	}

	if payload.TeamAccessPolicies != nil && !reflect.DeepEqual(payload.TeamAccessPolicies, endpointGroup.TeamAccessPolicies) {
		endpointGroup.TeamAccessPolicies = payload.TeamAccessPolicies
		updateAuthorizations = true
	}

	if updateAuthorizations {
		endpoints, err := tx.Endpoint().Endpoints()
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve environments from the database", err)
		}

		for _, endpoint := range endpoints {
			if endpoint.GroupID == endpointGroup.ID {
				if endpoint.Type == portaineree.KubernetesLocalEnvironment || endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
					err = handler.AuthorizationService.CleanNAPWithOverridePolicies(tx, &endpoint, endpointGroup)
					if err != nil {
						// Update flag with endpoint and continue
						go func(endpointID portaineree.EndpointID, endpointGroupID portaineree.EndpointGroupID) {
							err := handler.PendingActionsService.Create(portaineree.PendingActions{
								EndpointID: endpointID,
								Action:     "CleanNAPWithOverridePolicies",
								ActionData: endpointGroupID,
							})
							if err != nil {
								log.Error().Err(err).Msgf("Unable to create pending action to clean NAP with override policies for endpoint (%d) and endpoint group (%d).", endpointID, endpointGroupID)
							}
						}(endpoint.ID, endpointGroup.ID)
					}
				}
			}
		}
	}

	err = tx.EndpointGroup().Update(endpointGroup.ID, endpointGroup)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to persist environment group changes inside the database", err)
	}

	if updateAuthorizations {
		err = handler.AuthorizationService.UpdateUsersAuthorizationsTx(tx)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to update user authorizations", err)
		}
	}

	if tagsChanged {
		endpoints, err := tx.Endpoint().Endpoints()
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve environments from the database", err)
		}

		for _, endpoint := range endpoints {
			if endpoint.GroupID == endpointGroup.ID {
				err = handler.updateEndpointRelations(tx, &endpoint, endpointGroup)
				if err != nil {
					return nil, httperror.InternalServerError("Unable to persist environment relations changes inside the database", err)
				}
			}
		}
	}

	return endpointGroup, nil
}
