package gitops

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type repositoryReferenceListPayload struct {
	Repository      string `validate:"required" json:"repository"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	GitCredentialID int    `json:"gitCredentialID"`
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

	repositoryUsername, repositoryPassword, httpErr := handler.extractGitCredential(payload.Username, payload.Password, payload.GitCredentialID)
	if httpErr != nil {
		return httpErr
	}

	refs, err := handler.GitService.ListRefs(payload.Repository, repositoryUsername, repositoryPassword, hardRefresh)
	if err != nil {
		return httperror.InternalServerError("Git returned an error for listing refs", err)
	}

	return response.JSON(w, refs)
}
