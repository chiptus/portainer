package stacks

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/portainer/libhttp/response"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/stacks/deployments"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"
)

// @id WebhookInvoke
// @summary Webhook for triggering stack updates from git
// @description **Access policy**: public
// @tags stacks
// @param webhookID path string true "Stack identifier"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 409 "Conflict"
// @failure 500 "Server error"
// @router /stacks/webhooks/{webhookID} [post]
func (handler *Handler) webhookInvoke(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	webhookID, err := retrieveUUIDRouteVariableValue(r, "webhookID")
	if err != nil {
		return httperror.BadRequest("Invalid webhook identifier route variable", err)
	}

	stack, err := handler.DataStore.Stack().StackByWebhookID(webhookID.String())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if handler.DataStore.IsErrObjectNotFound(err) {
			statusCode = http.StatusNotFound
		}

		return &httperror.HandlerError{StatusCode: statusCode, Message: "Unable to find the stack by webhook ID", Err: err}
	}

	forcePullImage, envs := parseQuery(r.URL.Query())

	if err = deployments.RedeployWhenChanged(stack.ID, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService, envs, forcePullImage); err != nil {
		if _, ok := err.(*deployments.StackAuthorMissingErr); ok {
			return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: "Autoupdate for the stack isn't available", Err: err}
		}

		log.Error().Err(err).Msg("failed to update the stack")

		return httperror.InternalServerError("Failed to update the stack", err)
	}

	return response.Empty(w)
}

func parseQuery(query url.Values) (*bool, []portaineree.Pair) {
	var forcePullImage *bool
	envs := []portaineree.Pair{}
	for key, value := range query {
		val := value[len(value)-1]

		if key == "pullimage" {
			v, err := strconv.ParseBool(val)
			if err == nil {
				forcePullImage = &v
			}
			continue
		}

		envs = append(envs, portaineree.Pair{Name: key, Value: val})
	}

	return forcePullImage, envs
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
