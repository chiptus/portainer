package edgestacks

import (
	"fmt"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer/api/git/update"
)

// gitAutoUpdate checks if the git repository has changed and updates the stack if needed
func (handler *Handler) gitAutoUpdate(edgeStackId portaineree.EdgeStackID) error {
	return handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		edgeStack, err := tx.EdgeStack().EdgeStack(edgeStackId)
		if err != nil {
			return errors.WithMessage(err, "failed to find edge stack")
		}

		// if the stack is not using git, we force redeploy
		if edgeStack.GitConfig == nil {
			return tx.EdgeStack().UpdateEdgeStackFunc(edgeStackId, func(edgeStack *portaineree.EdgeStack) {
				edgeStack.Version = edgeStack.Version + 1
			})
		}

		updated, newHash, err := update.UpdateGitObject(
			handler.GitService,
			fmt.Sprintf("edge_stack:%d", edgeStack.ID),
			edgeStack.GitConfig,
			edgeStack.AutoUpdate != nil && edgeStack.AutoUpdate.ForceUpdate,
			edgeStack.ProjectPath)
		if err != nil {
			return err
		}

		if !updated {
			return nil
		}

		err = tx.EdgeStack().UpdateEdgeStackFunc(edgeStackId, func(edgeStack *portaineree.EdgeStack) {
			edgeStack.GitConfig.ConfigHash = newHash
			edgeStack.Version = edgeStack.Version + 1
		})

		if err != nil {
			return errors.WithMessage(err, "failed updating edge stack")
		}

		return nil
	})

}
