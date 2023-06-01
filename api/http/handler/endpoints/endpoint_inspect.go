package endpoints

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"
	"github.com/rs/zerolog/log"
)

// @id EndpointInspect
// @summary Inspect an environment(endpoint)
// @description Retrieve details about an environment(endpoint).
// @description **Access policy**: restricted
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id} [get]
func (handler *Handler) endpointInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, false)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	hideFields(endpoint)
	endpointutils.UpdateEdgeEndpointHeartbeat(endpoint, settings)
	endpoint.ComposeSyntaxMaxVersion = handler.ComposeStackManager.ComposeSyntaxMaxVersion()

	if !excludeSnapshot(r) {
		err = handler.SnapshotService.FillSnapshotData(endpoint)
		if err != nil {
			return httperror.InternalServerError("Unable to add snapshot data", err)
		}
	}

	if endpointutils.IsKubernetesEndpoint(endpoint) {
		isServerMetricsDetected := endpoint.Kubernetes.Flags.IsServerMetricsDetected
		if !isServerMetricsDetected && handler.K8sClientFactory != nil {
			endpointutils.InitialMetricsDetection(
				endpoint,
				handler.DataStore.Endpoint(),
				handler.K8sClientFactory,
			)
		}

		isServerStorageDetected := endpoint.Kubernetes.Flags.IsServerStorageDetected
		if !isServerStorageDetected && handler.K8sClientFactory != nil {
			endpointutils.InitialStorageDetection(
				endpoint,
				handler.DataStore.Endpoint(),
				handler.K8sClientFactory,
			)
		}

		existingRule, err := handler.DataStore.PodSecurity().PodSecurityByEndpointID(int(endpoint.ID))
		if err == nil {
			// Upgrade the gatekeeper if needed
			isGateKeeperRequireUpgrade := endpoint.PostInitMigrations.MigrateGateKeeper
			if isGateKeeperRequireUpgrade {
				gateKeeper := podsecurity.NewGateKeeper(
					handler.KubernetesDeployer,
					handler.AssetsPath,
				)

				kubeclient, err := handler.K8sClientFactory.GetKubeClient(endpoint)
				if err != nil {
					log.Error().Msgf("Error creating kubeclient for endpoint: %d", endpoint.ID)
				} else {
					_, err = kubeclient.GetNamespaces()
					if err != nil {
						log.Error().Msgf("Updating GateKeeper. error connecting endpoint (%d): %s", endpoint.ID, err)
					} else {

						cli, err := handler.K8sClientFactory.CreateClient(endpoint)
						if err != nil {
							log.Error().Msgf("Updating GateKeeper. error creating clientset (%d): %s", endpoint.ID, err)
						} else {
							err = gateKeeper.UpgradeEndpoint(1, endpoint, kubeclient, cli, existingRule)
							if err != nil {
								log.Error().Msgf("Error updating GateKeeper for endpoint (%d): %s", endpoint.ID, err)
							}
						}
					}
				}

				endpoint.PostInitMigrations.MigrateGateKeeper = false
				err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
				if err != nil {
					log.Error().Msgf("Error setting MigrateGateKeeper flag for endpoint %d : %s", endpoint.ID, err)
				}
			}
		}
	}

	return response.JSON(w, endpoint)
}

func excludeSnapshot(r *http.Request) bool {
	excludeSnapshot, _ := request.RetrieveBooleanQueryParameter(r, "excludeSnapshot", true)

	return excludeSnapshot
}
