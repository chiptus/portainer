package stacks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	consts "github.com/portainer/portainer-ee/api/useractivity"
	log "github.com/sirupsen/logrus"
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

// RedeployWhenChanged pull and redeploy the stack when  git repo changed
func RedeployWhenChanged(stackID portaineree.StackID, deployer StackDeployer, datastore dataservices.DataStore, gitService portaineree.GitService, activityService portaineree.UserActivityService) error {
	logger := log.WithFields(log.Fields{"stackID": stackID})
	logger.Debug("redeploying stack")

	stack, err := datastore.Stack().Stack(stackID)
	if err != nil {
		return errors.WithMessagef(err, "failed to get the stack %v", stackID)
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

		logger.Debug("not in update window. ignoring changes/webhooks")
		return nil // do nothing right now as we're not within the update window
	}

	author := stack.UpdatedBy
	if author == "" {
		author = stack.CreatedBy
	}

	user, err := datastore.User().UserByUsername(author)
	if err != nil {
		logger.WithFields(log.Fields{"author": author, "stack": stack.Name, "endpointID": stack.EndpointID}).Warn("cannot autoupdate a stack, stack author user is missing")
		return &StackAuthorMissingErr{int(stack.ID), author}
	}

	var gitCommitChangedOrForceUpdate bool
	if stack.GitConfig != nil && !stack.FromAppTemplate {
		logger.Debugln("The stack has a git config, try to poll from git repository.")
		username, password := "", ""
		if stack.GitConfig.Authentication != nil {
			username, password = stack.GitConfig.Authentication.Username, stack.GitConfig.Authentication.Password
		}
		newHash, err := gitService.LatestCommitID(stack.GitConfig.URL, stack.GitConfig.ReferenceName, username, password)
		if err != nil {
			return errors.WithMessagef(err, "failed to fetch latest commit id of the stack %v", stack.ID)
		}

		if !strings.EqualFold(newHash, string(stack.GitConfig.ConfigHash)) || (stack.AutoUpdate != nil && stack.AutoUpdate.ForceUpdate) {
			cloneParams := &cloneRepositoryParameters{
				url:   stack.GitConfig.URL,
				ref:   stack.GitConfig.ReferenceName,
				toDir: stack.ProjectPath,
			}
			if stack.GitConfig.Authentication != nil {
				cloneParams.auth = &gitAuth{
					username: username,
					password: password,
				}
			}

			if err := cloneGitRepository(gitService, cloneParams); err != nil {
				return errors.WithMessagef(err, "failed to do a fresh clone of the stack %v", stack.ID)
			}
			stack.UpdateDate = time.Now().Unix()
			stack.GitConfig.ConfigHash = newHash
			gitCommitChangedOrForceUpdate = true
		}
	}
	forcePullImage := stack.AutoUpdate == nil || stack.AutoUpdate.ForcePullImage
	if !forcePullImage && !gitCommitChangedOrForceUpdate {
		return nil
	}

	registries, err := getUserRegistries(datastore, user, endpoint.ID)
	if err != nil {
		return err
	}

	switch stack.Type {
	case portaineree.DockerComposeStack:
		logger.Debugf("Compose stack redeploy with pull image flag: %t", forcePullImage)
		err := deployer.DeployComposeStack(stack, endpoint, registries, forcePullImage, stack.AutoUpdate != nil && stack.AutoUpdate.ForceUpdate)
		if err != nil {
			return errors.WithMessagef(err, "failed to deploy a docker compose stack %v", stackID)
		}
	case portaineree.DockerSwarmStack:
		err := deployer.DeploySwarmStack(stack, endpoint, registries, true, true)
		if err != nil {
			return errors.WithMessagef(err, "failed to deploy a docker compose stack %v", stackID)
		}
	case portaineree.KubernetesStack:
		logger.Debugf("deploying a kube app")
		err := deployer.DeployKubernetesStack(stack, endpoint, user)
		if err != nil {
			return errors.WithMessagef(err, "failed to deploy a kubternetes app stack %v", stackID)
		}
	default:
		return errors.Errorf("cannot update stack, type %v is unsupported", stack.Type)
	}

	if err := datastore.Stack().UpdateStack(stack.ID, stack); err != nil {
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

func getUserRegistries(datastore dataservices.DataStore, user *portaineree.User, endpointID portaineree.EndpointID) ([]portaineree.Registry, error) {
	registries, err := datastore.Registry().Registries()
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

type cloneRepositoryParameters struct {
	url   string
	ref   string
	toDir string
	auth  *gitAuth
}

type gitAuth struct {
	username string
	password string
}

func cloneGitRepository(gitService portaineree.GitService, cloneParams *cloneRepositoryParameters) error {
	if cloneParams.auth != nil {
		return gitService.CloneRepository(cloneParams.toDir, cloneParams.url, cloneParams.ref, cloneParams.auth.username, cloneParams.auth.password)
	}
	return gitService.CloneRepository(cloneParams.toDir, cloneParams.url, cloneParams.ref, "", "")
}
