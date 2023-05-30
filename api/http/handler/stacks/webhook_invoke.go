package stacks

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/portainer/libhttp/response"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/stacks/deployments"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"

	"github.com/gofrs/uuid"
)

// @id StacksWebhookInvoke
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

	params, err := parseQuery(r.URL.Query())
	if err != nil {
		return httperror.BadRequest("Invalid query string", err)
	}

	if err = deployments.RedeployWhenChanged(stack.ID, handler.StackDeployer, handler.DataStore, handler.GitService, handler.userActivityService, params); err != nil {
		var StackAuthorMissingErr *deployments.StackAuthorMissingErr
		if errors.As(err, &StackAuthorMissingErr) {
			return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: "Autoupdate for the stack isn't available", Err: err}
		}

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

func parseQuery(query url.Values) (*deployments.RedeployOptions, error) {
	options := &deployments.RedeployOptions{}

	options.AdditionalEnvVars = make([]portaineree.Pair, 0)
	for key, value := range query {
		val := value[len(value)-1]

		switch key {
		case "pullimage":
			v, err := strconv.ParseBool(val)
			if err != nil {
				return nil, err
			}

			options.PullDockerImage = &v
		case "rollout-restart":
			switch val {
			case "all":
				options.RolloutRestartK8sAll = true
			case "":
				return nil, fmt.Errorf("rollout-restart value cannot be empty")
			default:
				options.RolloutRestartK8sResourceList = strings.Split(val, ",")
			}

		default:
			options.AdditionalEnvVars = append(options.AdditionalEnvVars, portaineree.Pair{Name: key, Value: val})
		}

	}

	return options, nil
}
