package gitops

import (
	"errors"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/asaskevich/govalidator"
)

type repositoryReferenceListPayload struct {
	Repository      string            `validate:"required" json:"repository"`
	Username        string            `json:"username"`
	Password        string            `json:"password"`
	StackID         portainer.StackID `json:"stackID"`
	GitCredentialID int               `json:"gitCredentialID"`
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

	var repositoryUsername, repositoryPassword string

	if payload.Password != "" || payload.GitCredentialID != 0 {
		username, password, httpErr := handler.extractGitCredential(payload.Username, payload.Password, payload.GitCredentialID)
		if httpErr != nil {
			return httpErr
		}

		repositoryUsername = username
		repositoryPassword = password
	}

	if payload.StackID != 0 && repositoryPassword == "" {
		stack, err := handler.dataStore.Stack().Read(portainer.StackID(payload.StackID))
		if err != nil {
			if handler.dataStore.IsErrObjectNotFound(err) {
				return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
			}
			return httperror.InternalServerError("Failed to locate the stack", err)
		}

		if stack.GitConfig != nil && stack.GitConfig.Authentication != nil {
			username, password, httpErr := handler.extractGitCredential(
				stack.GitConfig.Authentication.Username,
				stack.GitConfig.Authentication.Password,
				stack.GitConfig.Authentication.GitCredentialID,
			)
			if httpErr != nil {
				return httpErr
			}
			repositoryUsername = username
			repositoryPassword = password
		}
	}

	refs, err := handler.GitService.ListRefs(payload.Repository, repositoryUsername, repositoryPassword, hardRefresh, payload.TLSSkipVerify)
	if err != nil {
		return httperror.InternalServerError("Git returned an error for listing refs", err)
	}

	return response.JSON(w, refs)
}
