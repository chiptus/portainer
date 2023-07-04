package edgestacks

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	consts "github.com/portainer/portainer-ee/api/useractivity"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/git/update"
	"github.com/rs/zerolog/log"
)

// autoUpdate checks if the git repository or env vars have changed and updates the stack if needed
func (handler *Handler) autoUpdate(edgeStackId portaineree.EdgeStackID, envVars []portainer.Pair) error {
	err := handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		edgeStack, err := tx.EdgeStack().EdgeStack(edgeStackId)
		if err != nil {
			return errors.WithMessage(err, "failed to find edge stack")
		}

		gitUpdated := false
		oldHash := ""
		// if the stack is not using git, we force redeploy
		if edgeStack.GitConfig != nil {
			updated, newHash, err := update.UpdateGitObject(
				handler.GitService,
				fmt.Sprintf("edge_stack:%d", edgeStack.ID),
				edgeStack.GitConfig,
				edgeStack.AutoUpdate != nil && edgeStack.AutoUpdate.ForceUpdate,
				true,
				edgeStack.ProjectPath)
			if err != nil {
				return err
			}

			oldHash = edgeStack.GitConfig.ConfigHash
			edgeStack.GitConfig.ConfigHash = newHash

			gitUpdated = updated
		}

		envVarsUpdated, newEnvVars := upsertEnvVars(edgeStack.EnvVars, envVars)

		if !gitUpdated && !envVarsUpdated {
			return nil
		}

		edgeStack.EnvVars = newEnvVars
		var config []byte
		if edgeStack.GitConfig == nil {
			fileName := edgeStack.EntryPoint
			if edgeStack.DeploymentType == portaineree.EdgeStackDeploymentKubernetes {
				fileName = edgeStack.ManifestPath
			}

			projectPath := handler.FileService.FormProjectPathByVersion(edgeStack.ProjectPath, edgeStack.Version, "")
			stackFileContent, err := handler.FileService.GetFileContent(projectPath, fileName)
			if err != nil {
				log.Warn().
					Err(err).
					Int("stack_id", int(edgeStackId)).
					Msg("failed to get stack file content for stack")
			}

			config = stackFileContent
		}

		relationConfig, err := edge.FetchEndpointRelationsConfig(tx)
		if err != nil {
			return fmt.Errorf("unable to retrieve environments relations config from database: %w", err)
		}

		relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(edgeStack.EdgeGroups, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
		if err != nil {
			return fmt.Errorf("unable to retrieve edge stack related environments from database: %w", err)
		}

		err = handler.updateStackVersion(edgeStack, edgeStack.DeploymentType, config, oldHash, relatedEndpointIds)
		if err != nil {
			return fmt.Errorf("unable to update stack version: %w", err)
		}

		err = tx.EdgeStack().UpdateEdgeStack(edgeStackId, edgeStack)
		if err != nil {
			return fmt.Errorf("failed updating edge stack: %w", err)
		}

		for _, endpointID := range relatedEndpointIds {
			endpoint, err := tx.Endpoint().Endpoint(endpointID)
			if err != nil {
				return fmt.Errorf("unable to retrieve environment from the database: %w", err)
			}

			err = handler.edgeAsyncService.ReplaceStackCommandTx(tx, endpoint, edgeStack.ID)
			if err != nil {
				return fmt.Errorf("unable to store edge async command into the database: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return httperror.InternalServerError("failed in git auto update", err)

	}

	edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(edgeStackId)
	if err != nil {
		return httperror.InternalServerError("failed to find edge stack", err)
	}

	if handler.userActivityService != nil {
		if edgeStack.GitConfig != nil && edgeStack.GitConfig.Authentication != nil &&
			edgeStack.GitConfig.Authentication.Password != "" {
			edgeStack.GitConfig.Authentication.Password = consts.RedactedValue
		}

		body, _ := json.Marshal(edgeStack)
		handler.userActivityService.LogUserActivity("", "Portainer", "[Internal] Edge stack auto update", body)
	}

	return nil
}

func upsertEnvVars(oldEnvVars, envVars []portainer.Pair) (bool, []portainer.Pair) {
	updated := false
	for _, env := range envVars {
		exist := false
		for index, stackEnv := range oldEnvVars {
			if env.Name == stackEnv.Name {
				oldEnvVars[index] = env
				updated = true
				exist = true
				break
			}
		}

		if !exist {
			oldEnvVars = append(oldEnvVars, env)
			updated = true
		}
	}

	return updated, oldEnvVars
}
