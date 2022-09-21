package gitops

import (
	"errors"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type repositoryFileSearchPayload struct {
	Repository string `validate:"required" json:"repository"`
	// Specific Git repository reference. If empty, the reference ref/heads/main will be set by default
	Reference       string `json:"reference" example:"refs/heads/develop"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	GitCredentialID int    `json:"gitCredentialID"`
	// Partial or completed file name. If empty, all filenames with included extensions will be returned
	Keyword string `json:"keyword" example:"docker-compose"`
	// Allow to provide specific file extension as the search result. If empty, the file extensions yml,yaml,hcl,json will be set by default
	Include string `json:"include" example:"json,yml"`
}

func (payload *repositoryFileSearchPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Repository) || !govalidator.IsURL(payload.Repository) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}

	return nil
}

// @id GitOperationRepoFilesSearch
// @summary Search the file path from a git repository files with specified extensions
// @description Search the file path from the git repository based on partial or completed filename
// @description **Access policy**: authenticated
// @tags gitops
// @security ApiKeyAuth
// @security jwt
// @param force query bool false "list the results without using cache"
// @param body body repositoryFileSearchPayload true "details"
// @success 200 {array} string "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /gitops/repo/files/search [post]
func (handler *Handler) gitOperationRepoFilesSearch(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload repositoryFileSearchPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	if payload.Reference == "" {
		payload.Reference = "refs/heads/main"
	}

	includedExtensions := []string{"yml", "yaml", "json", "hcl", "nomad"}
	if payload.Include != "" {
		includedExtensions = strings.Split(payload.Include, ",")
	}

	hardRefresh, _ := request.RetrieveBooleanQueryParameter(r, "force", true)

	repositoryUsername, repositoryPassword, httpErr := handler.extractGitCredential(payload.Username, payload.Password, payload.GitCredentialID)
	if httpErr != nil {
		return httpErr
	}

	files, err := handler.GitService.ListFiles(payload.Repository, payload.Reference, repositoryUsername, repositoryPassword, hardRefresh, includedExtensions)
	if err != nil {
		return httperror.InternalServerError("Git returned an error for listing files", err)
	}

	var ret []string
	if payload.Keyword == "" {
		return response.JSON(w, files)
	}

	for _, path := range files {
		if strings.Contains(path, payload.Keyword) {
			ret = append(ret, path)
		}
	}

	return response.JSON(w, ret)
}
