package edgestacks

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/git/update"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/edge/staggers"
	"github.com/portainer/portainer-ee/api/internal/set"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/git"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/rs/zerolog/log"
)

type stackGitUpdatePayload struct {
	GroupIds       []portaineree.EdgeGroupID
	DeploymentType *portaineree.EdgeStackDeploymentType
	AutoUpdate     *portaineree.AutoUpdateSettings
	RefName        string
	Authentication *gittypes.GitAuthentication
	// Update the stack file content from the git repository
	// If this is set to true, it indicates that the stack is being redeployed,
	// if it is false, it indicates that the stack is being updated
	UpdateVersion bool
	EnvVars       []portainer.Pair
	// List of Registries to use for this stack
	Registries []portaineree.RegistryID
	// Configuration for stagger updates
	StaggerConfig *portaineree.EdgeStaggerConfig
}

func (payload *stackGitUpdatePayload) Validate(r *http.Request) error {
	if err := update.ValidateAutoUpdateSettings(payload.AutoUpdate); err != nil {
		return err
	}

	if err := git.ValidateRepoAuthentication(payload.Authentication); err != nil {
		return err
	}

	if payload.GroupIds != nil && len(payload.GroupIds) == 0 {
		return httperrors.NewInvalidPayloadError("Invalid Edge group IDs. Must contain at least one Edge group ID")
	}

	return staggers.ValidateStaggerConfig(payload.StaggerConfig)
}

// @id edgeStackUpdateFromGit
// @summary Update git configuration and pulls the repository
// @description **Access policy**: authenticated
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Stack identifier"
// @param body body stackGitUpdatePayload true "Git configurations"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /edge_stacks/{id}/git [put]
func (handler *Handler) edgeStackUpdateFromGitHandler(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	payload, err := request.GetPayload[stackGitUpdatePayload](r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	user, err := handler.DataStore.User().Read(securityContext.UserID)
	if err != nil {
		return httperror.BadRequest("Cannot find context user", errors.Wrap(err, "failed to fetch the user"))
	}

	edgeStack, err := middlewares.FetchItem[portaineree.EdgeStack](r, contextKey)
	if err != nil {
		return httperror.BadRequest("Failed to fetch Edge stack from context", err)
	}

	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {

		relationConfig, err := edge.FetchEndpointRelationsConfig(tx)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environments relations config from database", err)
		}

		relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(edgeStack.EdgeGroups, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve edge stack related environments from database", err)
		}

		endpointsToAdd := set.Set[portaineree.EndpointID]{}
		groupIds := edgeStack.EdgeGroups
		if payload.GroupIds != nil {
			newRelated, newEndpoints, err := handler.handleChangeEdgeGroups(tx, edgeStack.ID, payload.GroupIds, relatedEndpointIds, relationConfig)
			if err != nil {
				return httperror.InternalServerError("Unable to handle edge groups change", err)
			}

			groupIds = payload.GroupIds

			relatedEndpointIds = newRelated
			endpointsToAdd = newEndpoints
		}

		auth, hErr := parseGitCredentials(tx, payload.Authentication, edgeStack.GitConfig.Authentication, user.ID)
		if hErr != nil {
			return hErr
		}

		// make a copy of the current commit hash before updating the git config
		oldCommitHash := ""
		if edgeStack.GitConfig != nil {
			oldCommitHash = edgeStack.GitConfig.ConfigHash
		}

		gitConfig := edgeStack.GitConfig
		err = handler.updateGitSettings(gitConfig, payload.RefName, auth, payload.UpdateVersion)
		if err != nil {
			return httperror.InternalServerError("Failed updating git settings", err)
		}

		jobId := ""
		if edgeStack.AutoUpdate != nil {
			jobId = edgeStack.AutoUpdate.JobID
		}

		updateSettings, err := handler.updateAutoUpdateSettings(edgeStack.ID, payload.AutoUpdate, jobId)
		if err != nil {
			return httperror.InternalServerError("Failed updating auto update settings", err)
		}

		username, password := extractGitCredentials(auth)

		if payload.UpdateVersion {
			projectPath := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, 0, gitConfig.ConfigHash)
			err = handler.GitService.CloneRepository(projectPath, gitConfig.URL, payload.RefName, username, password, gitConfig.TLSSkipVerify)
			if err != nil {
				return httperror.InternalServerError("Failed cloning repository", err)
			}
		}

		edgeStack.GitConfig = gitConfig
		edgeStack.EdgeGroups = groupIds
		edgeStack.EnvVars = payload.EnvVars

		if payload.Registries != nil {
			edgeStack.Registries = payload.Registries
		}

		if payload.DeploymentType != nil {
			edgeStack.DeploymentType = *payload.DeploymentType
		}

		edgeStack.AutoUpdate = updateSettings
		edgeStack.NumDeployments = len(relatedEndpointIds)
		edgeStack.StaggerConfig = payload.StaggerConfig

		if payload.UpdateVersion {
			var rollbackTo *int = nil // update git API doesn't support rollback now (v2.19.0)
			err = handler.updateStackVersion(edgeStack, edgeStack.DeploymentType, nil, oldCommitHash, relatedEndpointIds, rollbackTo)
			if err != nil {
				return fmt.Errorf("unable to update stack version: %w", err)
			}

			if edgeStack.StaggerConfig != nil && edgeStack.StaggerConfig.StaggerOption != portaineree.EdgeStaggerOptionAllAtOnce {
				if oldCommitHash == edgeStack.GitConfig.ConfigHash {
					log.Info().Msg("Stack file version has not changed")
				}
				// User may update the env vars, so we still need to redeploy the stack
				err = handler.staggerService.AddStaggerConfig(edgeStack.ID,
					edgeStack.StackFileVersion,
					edgeStack.StaggerConfig)
				if err != nil {
					return httperror.InternalServerError("Unable to update git edge stack", err)
				}
			}
		}

		if edgeStack.GitConfig != nil && edgeStack.GitConfig.Authentication != nil &&
			edgeStack.GitConfig.Authentication.GitCredentialID != 0 {
			// prevent the username and password from saving into db if the git
			// credential is used
			edgeStack.GitConfig.Authentication.Username = ""
			edgeStack.GitConfig.Authentication.Password = ""
		}

		err = tx.EdgeStack().UpdateEdgeStack(edgeStack.ID, edgeStack, payload.UpdateVersion)
		if err != nil {
			return httperror.InternalServerError("Failed updating edge stack", err)
		}

		if payload.UpdateVersion {
			// Stagger configuration check
			if handler.staggerService != nil &&
				payload.StaggerConfig != nil &&
				payload.StaggerConfig.StaggerOption != portaineree.EdgeStaggerOptionAllAtOnce {
				go handler.staggerService.StartStaggerJobForAsyncUpdate(edgeStack.ID,
					relatedEndpointIds,
					endpointsToAdd,
					edgeStack.StackFileVersion)

				return nil
			}

			for _, endpointID := range relatedEndpointIds {
				endpoint, err := tx.Endpoint().Endpoint(endpointID)
				if err != nil {
					return httperror.InternalServerError("Unable to retrieve environment from the database", err)
				}

				if !endpointsToAdd[endpoint.ID] {
					err = handler.edgeAsyncService.ReplaceStackCommandTx(tx, endpoint, edgeStack.ID)
					if err != nil {
						return httperror.InternalServerError("Unable to store edge async command into the database", err)
					}
				}
			}
		}

		return nil
	})

	return httperrors.TxResponse(err, func() *httperror.HandlerError {
		return response.Empty(w)
	})

}

func (handler *Handler) updateAutoUpdateSettings(edgeStackID portaineree.EdgeStackID, settings *portaineree.AutoUpdateSettings, oldJobID string) (*portaineree.AutoUpdateSettings, error) {
	// stop the auto update job if there is any
	if oldJobID != "" {
		err := handler.scheduler.StopJob(oldJobID)
		if err != nil {
			return nil, errors.WithMessage(err, "Failed stopping auto update job")
		}
	}

	if settings == nil {
		return nil, nil
	}

	jobID, err := handler.handleAutoUpdate(edgeStackID, settings)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed setting auto update")
	}

	settings.JobID = jobID

	return settings, nil
}

func extractGitCredentials(auth *gittypes.GitAuthentication) (username, password string) {
	if auth == nil {
		return "", ""
	}

	return auth.Username, auth.Password
}

func (handler *Handler) updateGitSettings(originalGitConfig *gittypes.RepoConfig, newRefName string, auth *gittypes.GitAuthentication, updateHash bool) error {
	originalGitConfig.ReferenceName = newRefName

	originalGitConfig.Authentication = auth

	username, password := extractGitCredentials(auth)

	// This can be used to verify if the authentication is valid
	newHash, err := handler.GitService.LatestCommitID(originalGitConfig.URL, originalGitConfig.ReferenceName, username, password, originalGitConfig.TLSSkipVerify)
	if err != nil {
		return errors.WithMessage(err, "Unable to fetch git repository")
	}

	if updateHash {
		// When the updateHash is true, it means that the stack is being redeployed
		// In case case, we need to update the latest commit id used for determining
		// if the stack file version should be updated later
		originalGitConfig.ConfigHash = newHash
	}

	return nil
}

func parseGitCredentials(tx dataservices.DataStoreTx, authSettings, defaults *gittypes.GitAuthentication, userID portaineree.UserID) (*gittypes.GitAuthentication, *httperror.HandlerError) {
	if authSettings == nil {
		return nil, nil
	}

	if authSettings.GitCredentialID == 0 {
		if authSettings.Password == "" {
			return defaults, nil
		}

		return &gittypes.GitAuthentication{
			Username: authSettings.Username,
			Password: authSettings.Password,
		}, nil
	}

	credential, err := tx.GitCredential().Read(portaineree.GitCredentialID(authSettings.GitCredentialID))
	if err != nil {
		return nil, httperror.NotFound("Git credential not found", err)
	}

	if credential.UserID != userID {
		return nil, httperror.Forbidden("User do not match", err)
	}

	return &gittypes.GitAuthentication{
		Username:        credential.Username,
		Password:        credential.Password,
		GitCredentialID: int(credential.ID),
	}, nil
}
