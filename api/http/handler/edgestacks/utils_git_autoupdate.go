package edgestacks

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	consts "github.com/portainer/portainer-ee/api/useractivity"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/git/update"
)

// gitAutoUpdate checks if the git repository has changed and updates the stack if needed
func (handler *Handler) gitAutoUpdate(edgeStackId portaineree.EdgeStackID) error {
	err := handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		edgeStack, err := tx.EdgeStack().EdgeStack(edgeStackId)
		if err != nil {
			return errors.WithMessage(err, "failed to find edge stack")
		}

		// if the stack is not using git, we force redeploy
		if edgeStack.GitConfig == nil {
			return tx.EdgeStack().UpdateEdgeStackFunc(edgeStackId, func(edgeStack *portaineree.EdgeStack) {
				edgeStack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
					Version: edgeStack.Version,
				}
				edgeStack.Version = edgeStack.Version + 1
			})
		}

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

		if !updated {
			return nil
		}

		err = tx.EdgeStack().UpdateEdgeStackFunc(edgeStackId, func(edgeStack *portaineree.EdgeStack) {
			edgeStack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
				Version:    edgeStack.Version,
				ConfigHash: edgeStack.GitConfig.ConfigHash,
			}

			edgeStack.GitConfig.ConfigHash = newHash
			edgeStack.Version = edgeStack.Version + 1
		})

		if err != nil {
			return errors.WithMessage(err, "failed updating edge stack")
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
