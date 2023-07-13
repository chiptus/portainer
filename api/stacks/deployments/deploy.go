package deployments

import (
	"encoding/json"
	"fmt"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/stacks/stackutils"
	consts "github.com/portainer/portainer-ee/api/useractivity"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/git/update"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type StackAuthorMissingErr struct {
	stackID    int
	authorName string
}

func (e *StackAuthorMissingErr) Error() string {
	return fmt.Sprintf("stack's %v author %s is missing", e.stackID, e.authorName)
}

// Clock is an interface to determine the current time
type Clock interface {
	Now() time.Time
}

// RealClock is an implementation of a clock that returns the time in UTC
type realClockUTC struct{}

func (rc realClockUTC) Now() time.Time {
	return time.Now().UTC()
}

// updateAllowed returns true if AutoUpdateWindow.Enabled = false or
// AutoUpdateWindow.Enabled=true and the current UTC time is between StartTime and EndTime
// StartTime always begins BEFORE EndTime.  If EndTime is < StartTime then EndTime
// falls into the next day
func updateAllowed(endpoint *portaineree.Endpoint, clock Clock) (bool, error) {
	if !endpoint.ChangeWindow.Enabled {
		return true, nil
	}

	tw, err := NewTimeWindow(endpoint.ChangeWindow.StartTime, endpoint.ChangeWindow.EndTime)
	if err != nil {
		return false, errors.WithMessagef(err, "invalid time window")
	}

	return tw.Within(clock.Now()), nil
}

type RedeployOptions struct {
	AdditionalEnvVars             []portaineree.Pair
	PullDockerImage               *bool
	RolloutRestartK8sAll          bool
	RolloutRestartK8sResourceList []string
}

// RedeployWhenChanged pull and redeploy the stack when  git repo changed
func RedeployWhenChanged(stackID portaineree.StackID, deployer StackDeployer, datastore dataservices.DataStore, gitService portainer.GitService, activityService portaineree.UserActivityService, options *RedeployOptions) error {
	log.Debug().Int("stack_id", int(stackID)).Msg("redeploying stack")

	stack, err := datastore.Stack().Read(stackID)
	if err != nil {
		return errors.WithMessagef(err, "failed to get the stack %v", stackID)
	}

	additionalEnvUpserted := false
	if options != nil {
		for _, env := range options.AdditionalEnvVars {
			exist := false
			for index, stackEnv := range stack.Env {
				if env.Name == stackEnv.Name {
					stack.Env[index] = portaineree.Pair{
						Name:  env.Name,
						Value: env.Value,
					}
					exist = true
					additionalEnvUpserted = true
					break
				}
			}

			if !exist {
				stack.Env = append(stack.Env, portaineree.Pair{
					Name:  env.Name,
					Value: env.Value,
				})
				additionalEnvUpserted = true
			}
		}
	}

	endpoint, err := datastore.Endpoint().Endpoint(stack.EndpointID)
	if err != nil {
		return errors.WithMessagef(err, "failed to find the environment %v associated to the stack %v", stack.EndpointID, stack.ID)
	}

	var clock realClockUTC

	// return early if redeployment is not within change window and this feature is enabled
	if allowed, err := updateAllowed(endpoint, clock); !allowed {
		if err != nil {
			return errors.WithMessagef(err, "failed to parse the time stored in the portainer database")
		}

		log.Debug().Int("stack_id", int(stackID)).Msg("not in update window. ignoring changes/webhooks")

		return nil // do nothing right now as we're not within the update window
	}

	author := stack.UpdatedBy
	if author == "" {
		author = stack.CreatedBy
	}

	user, err := datastore.User().UserByUsername(author)
	if err != nil {
		log.Warn().
			Int("stack_id", int(stackID)).
			Str("author", author).
			Str("stack", stack.Name).
			Int("endpoint_id", int(stack.EndpointID)).
			Msg("cannot auto update a stack, stack author user is missing")

		return &StackAuthorMissingErr{int(stack.ID), author}
	}

	var gitCommitChangedOrForceUpdate bool
	if !stack.FromAppTemplate {
		updated, newHash, err := update.UpdateGitObject(gitService, fmt.Sprintf("stack:%d", stackID), stack.GitConfig, stack.AutoUpdate != nil && stack.AutoUpdate.ForceUpdate, true, stack.ProjectPath)
		if err != nil {
			return err
		}

		if updated {
			if stack.GitConfig.ConfigHash != newHash {
				// Only when the config hash changed, we need to update the PreviousDeploymentInfo
				stack.PreviousDeploymentInfo = &portainer.StackDeploymentInfo{
					ConfigHash:  stack.GitConfig.ConfigHash,
					FileVersion: stack.StackFileVersion,
				}
				stack.StackFileVersion++
				stack.GitConfig.ConfigHash = newHash
			}
			stack.UpdateDate = time.Now().Unix()
			gitCommitChangedOrForceUpdate = updated
		}
	}

	forcePullImage := func() bool {
		if options != nil && options.PullDockerImage != nil {
			return *options.PullDockerImage
		}

		return stack.AutoUpdate == nil || stack.AutoUpdate.ForcePullImage
	}()

	if !forcePullImage && !gitCommitChangedOrForceUpdate && !additionalEnvUpserted && !isRollingRestart(options) {
		return nil
	}

	registries, err := getUserRegistries(datastore, user, endpoint.ID)
	if err != nil {
		return err
	}

	switch stack.Type {
	case portaineree.DockerComposeStack:
		log.Debug().Int("stack_id", int(stackID)).Bool("force_pull_image", forcePullImage).Msg("compose stack redeploy with pull image flag")

		if stackutils.IsGitStack(stack) {
			err = deployer.DeployRemoteComposeStack(stack, endpoint, registries, forcePullImage, stack.AutoUpdate != nil && stack.AutoUpdate.ForceUpdate)
		} else {
			err = deployer.DeployComposeStack(stack, endpoint, registries, forcePullImage, stack.AutoUpdate != nil && stack.AutoUpdate.ForceUpdate)
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to deploy a docker compose stack %v", stackID)
		}
	case portaineree.DockerSwarmStack:
		if stackutils.IsGitStack(stack) {
			err = deployer.DeployRemoteSwarmStack(stack, endpoint, registries, true, forcePullImage)
		} else {
			err = deployer.DeploySwarmStack(stack, endpoint, registries, true, forcePullImage)
		}
		if err != nil {
			return errors.WithMessagef(err, "failed to deploy a docker compose stack %v", stackID)
		}
	case portaineree.KubernetesStack:
		log.Debug().
			Int("stack_id", int(stackID)).
			Msg("deploying a kube app")

		action := "restart"
		if options != nil && options.RolloutRestartK8sAll {
			err = deployer.RestartKubernetesStack(stack, endpoint, user, nil)
		} else if options != nil && len(options.RolloutRestartK8sResourceList) != 0 {
			err = deployer.RestartKubernetesStack(stack, endpoint, user, options.RolloutRestartK8sResourceList)
		} else {
			err = deployer.DeployKubernetesStack(stack, endpoint, user)
			action = "deploy"
		}

		if err != nil {
			return errors.WithMessagef(err, "failed to %s a kubternetes app stack %v", action, stackID)
		}
	default:
		return errors.Errorf("cannot update stack, type %v is unsupported", stack.Type)
	}

	if err := datastore.Stack().Update(stack.ID, stack); err != nil {
		return errors.WithMessagef(err, "failed to update the stack %v", stack.ID)
	}

	if activityService != nil {
		if stack.GitConfig != nil && stack.GitConfig.Authentication != nil && stack.GitConfig.Authentication.Password != "" {
			stack.GitConfig.Authentication.Password = consts.RedactedValue
		}

		body, _ := json.Marshal(stack)
		activityService.LogUserActivity(author, endpoint.Name, "[INTERNAL] stack auto update", body)
	}

	return nil
}

func isRollingRestart(options *RedeployOptions) bool {
	if options == nil {
		return false
	}

	if options.RolloutRestartK8sAll || len(options.RolloutRestartK8sResourceList) > 0 {
		return true
	}

	return false
}

func getUserRegistries(datastore dataservices.DataStore, user *portaineree.User, endpointID portaineree.EndpointID) ([]portaineree.Registry, error) {
	registries, err := datastore.Registry().ReadAll()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve registries from the database")
	}

	if user.Role == portaineree.AdministratorRole {
		return registries, nil
	}

	userMemberships, err := datastore.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to fetch memberships of the stack author [%s]", user.Username)
	}

	filteredRegistries := make([]portaineree.Registry, 0, len(registries))
	for _, registry := range registries {
		if security.AuthorizedRegistryAccess(&registry, user, userMemberships, endpointID) {
			filteredRegistries = append(filteredRegistries, registry)
		}
	}

	return filteredRegistries, nil
}
