package edgestacks

import (
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	eefs "github.com/portainer/portainer-ee/api/filesystem"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
)

type edgeStackFromGitRepositoryPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// URL of a Git repository hosting the Stack file
	RepositoryURL string `example:"https://github.com/openfaas/faas" validate:"required"`
	// Reference name of a Git repository hosting the Stack file
	RepositoryReferenceName string `example:"refs/heads/master"`
	// Use basic authentication to clone the Git repository
	RepositoryAuthentication bool `example:"true"`
	// Username used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryUsername string `example:"myGitUsername"`
	// Password used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryPassword string `example:"myGitPassword"`
	// GitCredentialID used to identify the binded git credential
	RepositoryGitCredentialID portaineree.GitCredentialID `example:"0"`
	// Path to the Stack file inside the Git repository
	FilePathInRepository string `example:"docker-compose.yml" default:"docker-compose.yml"`
	// List of identifiers of EdgeGroups
	EdgeGroups []portaineree.EdgeGroupID `example:"1"`
	// Deployment type to deploy this stack
	// Valid values are: 0 - 'compose', 1 - 'kubernetes', 2 - 'nomad'
	// compose is enabled only for docker environments
	// kubernetes is enabled only for kubernetes environments
	// nomad is enabled only for nomad environments
	DeploymentType portaineree.EdgeStackDeploymentType `example:"0" enums:"0,1,2"`
	// List of Registries to use for this stack
	Registries []portaineree.RegistryID
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
	// Pre Pull image
	PrePullImage bool `example:"false"`
	// Retry deploy
	RetryDeploy bool `example:"false"`
	// TLSSkipVerify skips SSL verification when cloning the Git repository
	TLSSkipVerify bool `example:"false"`
}

func (payload *edgeStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return httperrors.NewInvalidPayloadError("Invalid stack name")
	}

	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return httperrors.NewInvalidPayloadError("Invalid repository URL. Must correspond to a valid URL format")
	}

	if payload.RepositoryAuthentication && govalidator.IsNull(payload.RepositoryPassword) && payload.RepositoryGitCredentialID == 0 {
		return httperrors.NewInvalidPayloadError("Invalid repository credentials. Password must be specified when authentication is enabled")
	}

	if payload.DeploymentType != portaineree.EdgeStackDeploymentCompose && payload.DeploymentType != portaineree.EdgeStackDeploymentKubernetes && payload.DeploymentType != portaineree.EdgeStackDeploymentNomad {
		return httperrors.NewInvalidPayloadError("Invalid deployment type")
	}

	if govalidator.IsNull(payload.FilePathInRepository) {
		switch payload.DeploymentType {
		case portaineree.EdgeStackDeploymentCompose:
			payload.FilePathInRepository = filesystem.ComposeFileDefaultName
		case portaineree.EdgeStackDeploymentKubernetes:
			payload.FilePathInRepository = filesystem.ManifestFileDefaultName
		case portaineree.EdgeStackDeploymentNomad:
			payload.FilePathInRepository = eefs.NomadJobFileDefaultName
		}
	}

	if len(payload.EdgeGroups) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid edge groups. At least one edge group must be specified")
	}

	return nil
}

// @id EdgeStackCreateRepository
// @summary Create an EdgeStack from a git repository
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param method query string true "Creation Method" Enums(file,string,repository)
// @param body body edgeStackFromGitRepositoryPayload true "stack config"
// @param dryrun query string false "if true, will not create an edge stack, but just will check the settings and return a non-persisted edge stack object"
// @success 200 {object} portaineree.EdgeStack
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/create/repository [post]
func (handler *Handler) createEdgeStackFromGitRepository(r *http.Request, dryrun bool, userID portaineree.UserID) (*portaineree.EdgeStack, error) {
	var payload edgeStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	stack, err := handler.edgeStacksService.BuildEdgeStack(payload.Name, payload.DeploymentType, payload.EdgeGroups, payload.Registries, "", payload.UseManifestNamespaces, payload.PrePullImage, false, payload.RetryDeploy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create edge stack object")
	}

	if dryrun {
		return stack, nil
	}

	repoConfig := gittypes.RepoConfig{
		URL:            payload.RepositoryURL,
		ReferenceName:  payload.RepositoryReferenceName,
		ConfigFilePath: payload.FilePathInRepository,
		TLSSkipVerify:  payload.TLSSkipVerify,
	}

	if payload.RepositoryAuthentication {
		repoConfig.Authentication = &gittypes.GitAuthentication{
			Username: payload.RepositoryUsername,
			Password: payload.RepositoryPassword,
		}
	}

	return handler.edgeStacksService.PersistEdgeStack(stack, func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (configPath string, manifestPath string, projectPath string, err error) {
		return handler.storeManifestFromGitRepository(stackFolder, relatedEndpointIds, payload.DeploymentType, userID, payload.RepositoryGitCredentialID, repoConfig)
	})
}

func (handler *Handler) storeManifestFromGitRepository(stackFolder string, relatedEndpointIds []portaineree.EndpointID, deploymentType portaineree.EdgeStackDeploymentType, currentUserID portaineree.UserID, gitCredentialId portaineree.GitCredentialID, repositoryConfig gittypes.RepoConfig) (composePath, manifestPath, projectPath string, err error) {
	hasWrongType, err := hasWrongEnvironmentType(handler.DataStore.Endpoint(), relatedEndpointIds, deploymentType)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to check for existence of non fitting environments: %w", err)
	}
	if hasWrongType {
		return "", "", "", fmt.Errorf("edge stack with config do not match the environment type")
	}

	projectPath = handler.FileService.GetEdgeStackProjectPath(stackFolder)
	repositoryUsername := ""
	repositoryPassword := ""
	if repositoryConfig.Authentication != nil {
		if gitCredentialId != 0 {
			credential, err := handler.DataStore.GitCredential().GetGitCredential(gitCredentialId)
			if err != nil {
				return "", "", "", fmt.Errorf("git credential not found: %w", err)
			}

			// When creating the stack with an existing git credential, the git credential must be owned by the calling user
			if credential.UserID != currentUserID {
				return "", "", "", fmt.Errorf("couldn't retrieve the git credential for another user: %w", err)
			}

			repositoryUsername = credential.Username
			repositoryPassword = credential.Password
		}

		if repositoryConfig.Authentication.Password != "" {
			repositoryUsername = repositoryConfig.Authentication.Username
			repositoryPassword = repositoryConfig.Authentication.Password
		}
	}

	err = handler.GitService.CloneRepository(projectPath, repositoryConfig.URL, repositoryConfig.ReferenceName, repositoryUsername, repositoryPassword, repositoryConfig.TLSSkipVerify)
	if err != nil {
		if errors.Is(err, gittypes.ErrAuthenticationFailure) {
			return "", "", "", errInvalidGitCredential
		}
		return "", "", "", err
	}

	if deploymentType == portaineree.EdgeStackDeploymentCompose {
		return repositoryConfig.ConfigFilePath, "", projectPath, nil
	}

	if deploymentType == portaineree.EdgeStackDeploymentKubernetes {
		return "", repositoryConfig.ConfigFilePath, projectPath, nil
	}

	if deploymentType == portaineree.EdgeStackDeploymentNomad {
		return repositoryConfig.ConfigFilePath, "", projectPath, nil
	}

	errMessage := fmt.Sprintf("unknown deployment type: %d", deploymentType)
	return "", "", "", httperrors.NewInvalidPayloadError(errMessage)
}
