package gitops

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type repositoryReferenceListPayload struct {
	Repository      string              `validate:"required" json:"repository"`
	Username        string              `json:"username"`
	Password        string              `json:"password"`
	StackID         portaineree.StackID `json:"stackID"`
	GitCredentialID int                 `json:"gitCredentialID"`
	// TLSSkipVerify skips SSL verification when cloning the Git repository
	TLSSkipVerify bool `example:"false"`
}

func (payload *repositoryReferenceListPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Repository) || !govalidator.IsURL(payload.Repository) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	return nil
}

// @id GitOperationRepoRefs
// @summary List the refs of a git repository
// @description List all the refs of a git repository
// @description Will return all refs of a git repository
// @description **Access policy**: authenticated
// @tags gitops
// @security ApiKeyAuth
// @security jwt
// @param force query bool false "list the results without using cache"
// @param body body repositoryReferenceListPayload true "details"
// @success 200 {array} string "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /gitops/repo/refs [post]
func (handler *Handler) gitOperationRepoRefs(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload repositoryReferenceListPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	hardRefresh, _ := request.RetrieveBooleanQueryParameter(r, "force", true)

	repositoryUsername := ""
	repositoryPassword := ""
	if payload.StackID != 0 {
		stack, err := handler.dataStore.Stack().Read(portaineree.StackID(payload.StackID))
		if handler.dataStore.IsErrObjectNotFound(err) {
			return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
		} else if err != nil {
			return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
		} else if stack.GitConfig == nil {
			msg := "No Git config in the found stack"
			return httperror.InternalServerError(msg, errors.New(msg))
		} else if stack.GitConfig.Authentication == nil {
			msg := "No Git credential in the found stack"
			return httperror.InternalServerError(msg, errors.New(msg))
		}

		repositoryUsername = stack.GitConfig.Authentication.Username
		repositoryPassword = stack.GitConfig.Authentication.Password

	} else {
		username, password, httpErr := handler.extractGitCredential(payload.Username, payload.Password, payload.GitCredentialID)
		if httpErr != nil {
			return httpErr
		}
		repositoryUsername = username
		repositoryPassword = password
	}

	refs, err := handler.GitService.ListRefs(payload.Repository, repositoryUsername, repositoryPassword, hardRefresh, payload.TLSSkipVerify)
	if err != nil {
		return httperror.InternalServerError("Git returned an error for listing refs", err)
	}

	return response.JSON(w, refs)
}
