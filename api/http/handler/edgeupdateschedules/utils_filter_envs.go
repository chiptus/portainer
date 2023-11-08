package edgeupdateschedules

import (
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/set"
	pslices "github.com/portainer/portainer-ee/api/internal/slices"
	portainer "github.com/portainer/portainer/api"
)

// filterEnvironments fetches all environments related to the edge group ids
// together with their version data
//
// it also filters out environments that are already updated and validates that all
// environments are of the same type
func (handler *Handler) filterEnvironments(tx dataservices.DataStoreTx, edgeGroupIds []portainer.EdgeGroupID, version string, rollback bool) ([]portainer.EndpointID, portainer.EndpointType, error) {
	relationConfig, err := edge.FetchEndpointRelationsConfig(handler.dataStore)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "unable to fetch environment relations config")
	}

	groupsEnvironmentsIds, err := edge.EdgeStackRelatedEndpoints(edgeGroupIds, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "unable to fetch related environments")
	}

	if len(groupsEnvironmentsIds) == 0 {
		return nil, 0, errors.New("no related environments")
	}

	environments, err := tx.Endpoint().Endpoints()
	if err != nil {
		return nil, 0, errors.WithMessage(err, "unable to fetch environments")
	}

	relatedEnvironmentIdsSet := set.ToSet(groupsEnvironmentsIds)

	relatedEnvironments := []portaineree.Endpoint{}

	semverConstraint, err := semver.NewConstraint("< " + version)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "unable to parse version constraint")
	}

	var envType portainer.EndpointType
	for _, environment := range environments {
		if !relatedEnvironmentIdsSet.Contains(environment.ID) {
			continue
		}

		if envType == 0 {
			envType = environment.Type
		}

		if environment.Type != envType {
			return nil, 0, errors.New("environment type is not unified")
		}

		if rollback {
			if environment.Agent.PreviousVersion != version {
				continue
			}

		} else {
			err := handler.isUpdateSupported(tx, &environment)
			if err != nil {
				return nil, 0, errors.WithMessage(err, "unable to validate environment")
			}

			if environment.Agent.Version != "" {
				agentVersion, err := semver.NewVersion(environment.Agent.Version)
				if err != nil {
					return nil, 0, errors.WithMessage(err, "unable to parse agent version")
				}

				if !semverConstraint.Check(agentVersion) {
					continue
				}
			}

		}

		relatedEnvironments = append(relatedEnvironments, environment)
	}

	relatedEnvIds := pslices.Map(relatedEnvironments, func(environment portaineree.Endpoint) portainer.EndpointID {
		return environment.ID
	})

	if len(relatedEnvIds) == 0 {
		return nil, 0, errors.New("no related environments that require update")
	}

	return relatedEnvIds, envType, nil
}

func (handler *Handler) isUpdateSupported(tx dataservices.DataStoreTx, environment *portaineree.Endpoint) error {
	if !endpointutils.IsEdgeEndpoint(environment) {
		return errors.New("environment is not an edge endpoint, this feature is limited to edge endpoints")
	}

	if endpointutils.IsNomadEndpoint(environment) {
		// Nomad does not need to check snapshot
		return nil
	}

	if endpointutils.IsDockerEndpoint(environment) {
		snapshot, err := tx.Snapshot().Read(environment.ID)
		if err != nil {
			// if snapshot is missing, we require a tunnel, which will fetch the snapshot on close
			handler.ReverseTunnelService.SetTunnelStatusToRequired(environment.ID)

			return errors.WithMessage(err, "unable to fetch snapshot, please try again later")
		}

		if snapshot.Docker == nil {
			// if snapshot is missing, we require a tunnel, which will fetch the snapshot on close
			handler.ReverseTunnelService.SetTunnelStatusToRequired(environment.ID)

			return errors.New("missing docker snapshot, please try again later")
		}

		if snapshot.Docker.Swarm {
			return errors.New("swarm is not supported")
		}

		return nil
	}

	return errors.New("environment is not a docker/nomad endpoint, this feature is limited to docker/nomad endpoints")
}
