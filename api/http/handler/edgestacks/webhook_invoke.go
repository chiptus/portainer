package edgestacks

import (
	"net/http"
	"net/url"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// @id EdgeStackWebhookInvoke
// @summary Webhook for triggering edge stack updates from git
// @description **Access policy**: public
// @tags stacks
// @param webhookID path string true "Stack identifier"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 409 "Conflict"
// @failure 500 "Server error"
// @router /edge_stacks/webhooks/{webhookID} [post]
func (handler *Handler) webhookInvoke(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	webhookID, err := retrieveUUIDRouteVariableValue(r, "webhookID")
	if err != nil {
		return httperror.BadRequest("Invalid webhook identifier route variable", err)
	}

	edgeStack, err := handler.edgeStackByWebhook(webhookID.String())
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve edge stack from the database", err)
	}

	if edgeStack == nil {
		return httperror.NotFound("Unable to find edge stack with the specified webhook id", nil)
	}

	envVars := parseQuery(r.URL.Query())

	if err = handler.autoUpdate(edgeStack.ID, envVars); err != nil {
		log.Error().Err(err).Msg("failed to update the stack")

		return httperror.InternalServerError("Failed to update the stack", err)
	}

	return response.Empty(w)
}

func retrieveUUIDRouteVariableValue(r *http.Request, name string) (uuid.UUID, error) {
	webhookID, err := request.RetrieveRouteVariableValue(r, name)
	if err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.FromString(webhookID)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func (handler *Handler) edgeStackByWebhook(webhookID string) (*portaineree.EdgeStack, error) {
	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to retrieve edge stacks from the database")
	}

	for i, stack := range edgeStacks {
		if strings.EqualFold(stack.Webhook, webhookID) {
			return &edgeStacks[i], nil
		}

		if stack.AutoUpdate != nil && strings.EqualFold(stack.AutoUpdate.Webhook, webhookID) {
			return &edgeStacks[i], nil
		}
	}

	return nil, nil
}

func parseQuery(query url.Values) []portainer.Pair {
	envVars := make([]portainer.Pair, 0)
	for key, value := range query {
		val := value[len(value)-1]

		envVars = append(envVars, portainer.Pair{Name: key, Value: val})
	}

	return envVars
}
