package edgestacks

import (
	"fmt"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/security"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

func (handler *Handler) edgeStackCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	method, err := request.RetrieveRouteVariableValue(r, "method")
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: method", err)
	}

	dryrun, _ := request.RetrieveBooleanQueryParameter(r, "dryrun", true)

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user details from authentication token", err)
	}

	var edgeStack *portaineree.EdgeStack
	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		edgeStack, err = handler.createSwarmStack(tx, method, dryrun, tokenData.ID, r)
		return err
	})
	if err != nil {
		switch {
		case httperrors.IsInvalidPayloadError(err):
			return httperror.BadRequest("Invalid payload", err)
		case httperrors.IsConflictError(err):
			return httperror.NewError(http.StatusConflict, err.Error(), err)
		default:
			return httperror.InternalServerError("Unable to create Edge stack", err)
		}
	}

	return response.JSON(w, edgeStack)
}

func (handler *Handler) createSwarmStack(tx dataservices.DataStoreTx, method string, dryrun bool, userID portaineree.UserID, r *http.Request) (*portaineree.EdgeStack, error) {

	switch method {
	case "string":
		return handler.createEdgeStackFromFileContent(r, tx, dryrun)
	case "repository":
		return handler.createEdgeStackFromGitRepository(r, tx, dryrun, userID)
	case "file":
		return handler.createEdgeStackFromFileUpload(r, tx, dryrun)
	}

	return nil, httperrors.NewInvalidPayloadError("Invalid value for query parameter: method. Value must be one of: string, repository or file")
}

// @id EdgeStackCreate
// @summary Create an EdgeStack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param method query string true "Creation Method" Enums(file,string,repository)
// @param body body object true "for body documentation see the relevant /edge_stacks/create/{method} endpoint"
// @success 200 {object} portaineree.EdgeStack
// @failure 500
// @failure 503 "Edge compute features are disabled"
// @deprecated
// @router /edge_stacks [post]
func deprecatedEdgeStackCreateUrlParser(w http.ResponseWriter, r *http.Request) (string, *httperror.HandlerError) {
	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return "", httperror.BadRequest("Invalid query parameter: method. Valid values are: file or string", err)
	}

	return fmt.Sprintf("/edge_stacks/create/%s", method), nil
}
