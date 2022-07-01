package endpoints

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/client"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/tag"
	portainer "github.com/portainer/portainer/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

type endpointUpdatePayload struct {
	// Name that will be used to identify this environment(endpoint)
	Name *string `example:"my-environment"`
	// URL or IP address of a Docker host
	URL *string `example:"docker.mydomain.tld:2375"`
	// URL or IP address where exposed containers will be reachable.\
	// Defaults to URL if not specified
	PublicURL *string `example:"docker.mydomain.tld:2375"`
	// Group identifier
	GroupID *int `example:"1"`
	// Require TLS to connect against this environment(endpoint)
	TLS *bool `example:"true"`
	// Skip server verification when using TLS
	TLSSkipVerify *bool `example:"false"`
	// Skip client verification when using TLS
	TLSSkipClientVerify *bool `example:"false"`
	// The status of the environment(endpoint) (1 - up, 2 - down)
	Status *int `example:"1"`
	// Azure application ID
	AzureApplicationID *string `example:"eag7cdo9-o09l-9i83-9dO9-f0b23oe78db4"`
	// Azure tenant ID
	AzureTenantID *string `example:"34ddc78d-4fel-2358-8cc1-df84c8o839f5"`
	// Azure authentication key
	AzureAuthenticationKey *string `example:"cOrXoK/1D35w8YQ8nH1/8ZGwzz45JIYD5jxHKXEQknk="`
	// List of tag identifiers to which this environment(endpoint) is associated
	TagIDs             []portaineree.TagID `example:"1,2"`
	UserAccessPolicies portaineree.UserAccessPolicies
	TeamAccessPolicies portaineree.TeamAccessPolicies
	// Associated Kubernetes data
	Kubernetes *portaineree.KubernetesData
	// Whether automatic update time restrictions are enabled
	ChangeWindow *portaineree.EndpointChangeWindow
	// The check in interval for edge agent (in seconds)
	EdgeCheckinInterval *int `example:"5"`

	Edge struct {
		// The ping interval for edge agent - used in edge async mode (in seconds)
		PingInterval *int `json:"PingInterval" example:"5"`
		// The snapshot interval for edge agent - used in edge async mode (in seconds)
		SnapshotInterval *int `json:"SnapshotInterval" example:"5"`
		// The command list interval for edge agent - used in edge async mode (in seconds)
		CommandInterval *int `json:"CommandInterval" example:"5"`
	}
}

func (payload *endpointUpdatePayload) Validate(r *http.Request) error {
	if payload.ChangeWindow != nil {
		err := validateAutoUpdateSettings(*payload.ChangeWindow)
		if err != nil {
			return errors.WithMessage(err, "Validation failed")
		}
	}

	return nil
}

// @id EndpointUpdate
// @summary Update an environment(endpoint)
// @description Update an environment(endpoint).
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags endpoints
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param body body endpointUpdatePayload true "Environment(Endpoint) details"
// @success 200 {object} portaineree.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id} [put]
func (handler *Handler) endpointUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	// make sure user has been issued a token
	permissionDeniedErr := "Permission denied to update environment"
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, permissionDeniedErr, err}
	}

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment identifier route variable", err}
	}

	var payload endpointUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an environment with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an environment with the specified identifier inside the database", err}
	}

	isAdmin := tokenData.Role == portaineree.AdministratorRole
	canK8sClusterSetup := isAdmin || false
	// if user is not a portainer admin, we might allow update k8s cluster config
	if !isAdmin {
		// check if the user can access cluster setup in the environment(endpoint) (environment admin)
		endpointRole, err := handler.AuthorizationService.GetUserEndpointRole(int(tokenData.ID), int(endpoint.ID))
		if err != nil {
			return &httperror.HandlerError{http.StatusForbidden, permissionDeniedErr, err}
		} else if !endpointRole.Authorizations[portaineree.OperationK8sClusterSetupRW] {
			err = errors.New(permissionDeniedErr)
			return &httperror.HandlerError{http.StatusForbidden, permissionDeniedErr, err}
		}
		// deny access if user can not access all namespaces
		if !endpointRole.Authorizations[portaineree.OperationK8sAccessAllNamespaces] {
			err = errors.New(permissionDeniedErr)
			return &httperror.HandlerError{http.StatusForbidden, permissionDeniedErr, err}
		} else {
			canK8sClusterSetup = true
		}
	}

	updateAuthorizations := false
	if canK8sClusterSetup && payload.Kubernetes != nil {
		endpoint.Kubernetes = *payload.Kubernetes

		if payload.Kubernetes.Configuration.RestrictDefaultNamespace !=
			endpoint.Kubernetes.Configuration.RestrictDefaultNamespace {
			updateAuthorizations = true
		}
	}

	groupIDChanged := false
	tagsChanged := false
	if isAdmin {
		if payload.Name != nil {
			name := *payload.Name
			isUnique, err := handler.isNameUnique(name, endpoint.ID)
			if err != nil {
				return httperror.InternalServerError("Unable to check if name is unique", err)
			}

			if !isUnique {
				return httperror.NewError(http.StatusConflict, "Name is not unique", nil)
			}

			endpoint.Name = name
		}

		if payload.URL != nil {
			endpoint.URL = *payload.URL
		}

		if payload.PublicURL != nil {
			endpoint.PublicURL = *payload.PublicURL
		}

		if payload.EdgeCheckinInterval != nil {
			endpoint.EdgeCheckinInterval = *payload.EdgeCheckinInterval
		}

		if payload.Edge.PingInterval != nil {
			endpoint.Edge.PingInterval = *payload.Edge.PingInterval
		}

		if payload.Edge.SnapshotInterval != nil {
			endpoint.Edge.SnapshotInterval = *payload.Edge.SnapshotInterval
		}

		if payload.Edge.CommandInterval != nil {
			endpoint.Edge.CommandInterval = *payload.Edge.CommandInterval
		}

		if payload.GroupID != nil {
			groupID := portaineree.EndpointGroupID(*payload.GroupID)
			groupIDChanged = groupID != endpoint.GroupID
			endpoint.GroupID = groupID
		}

		if payload.TagIDs != nil {
			payloadTagSet := tag.Set(payload.TagIDs)
			endpointTagSet := tag.Set((endpoint.TagIDs))
			union := tag.Union(payloadTagSet, endpointTagSet)
			intersection := tag.Intersection(payloadTagSet, endpointTagSet)
			tagsChanged = len(union) > len(intersection)

			if tagsChanged {
				removeTags := tag.Difference(endpointTagSet, payloadTagSet)

				for tagID := range removeTags {
					tag, err := handler.dataStore.Tag().Tag(tagID)
					if err != nil {
						return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a tag inside the database", err}
					}

					delete(tag.Endpoints, endpoint.ID)
					err = handler.dataStore.Tag().UpdateTag(tag.ID, tag)
					if err != nil {
						return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist tag changes inside the database", err}
					}
				}

				endpoint.TagIDs = payload.TagIDs
				for _, tagID := range payload.TagIDs {
					tag, err := handler.dataStore.Tag().Tag(tagID)
					if err != nil {
						return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a tag inside the database", err}
					}

					tag.Endpoints[endpoint.ID] = true

					err = handler.dataStore.Tag().UpdateTag(tag.ID, tag)
					if err != nil {
						return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist tag changes inside the database", err}
					}
				}
			}
		}

		if payload.UserAccessPolicies != nil && !reflect.DeepEqual(payload.UserAccessPolicies, endpoint.UserAccessPolicies) {
			updateAuthorizations = true
			endpoint.UserAccessPolicies = payload.UserAccessPolicies
		}

		if payload.TeamAccessPolicies != nil && !reflect.DeepEqual(payload.TeamAccessPolicies, endpoint.TeamAccessPolicies) {
			updateAuthorizations = true
			endpoint.TeamAccessPolicies = payload.TeamAccessPolicies
		}

		if payload.Status != nil {
			switch *payload.Status {
			case 1:
				endpoint.Status = portaineree.EndpointStatusUp
				break
			case 2:
				endpoint.Status = portaineree.EndpointStatusDown
				break
			default:
				break
			}
		}

		if endpoint.Type == portaineree.AzureEnvironment {
			credentials := endpoint.AzureCredentials
			if payload.AzureApplicationID != nil {
				credentials.ApplicationID = *payload.AzureApplicationID
			}
			if payload.AzureTenantID != nil {
				credentials.TenantID = *payload.AzureTenantID
			}
			if payload.AzureAuthenticationKey != nil {
				credentials.AuthenticationKey = *payload.AzureAuthenticationKey
			}

			httpClient := client.NewHTTPClient()
			_, authErr := httpClient.ExecuteAzureAuthenticationRequest(&credentials)
			if authErr != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to authenticate against Azure", authErr}
			}
			endpoint.AzureCredentials = credentials
		}

		if payload.TLS != nil {
			folder := strconv.Itoa(endpointID)

			if *payload.TLS {
				endpoint.TLSConfig.TLS = true
				if payload.TLSSkipVerify != nil {
					endpoint.TLSConfig.TLSSkipVerify = *payload.TLSSkipVerify

					if !*payload.TLSSkipVerify {
						caCertPath, _ := handler.FileService.GetPathForTLSFile(folder, portainer.TLSFileCA)
						endpoint.TLSConfig.TLSCACertPath = caCertPath
					} else {
						endpoint.TLSConfig.TLSCACertPath = ""
						handler.FileService.DeleteTLSFile(folder, portainer.TLSFileCA)
					}
				}

				if payload.TLSSkipClientVerify != nil {
					if !*payload.TLSSkipClientVerify {
						certPath, _ := handler.FileService.GetPathForTLSFile(folder, portainer.TLSFileCert)
						endpoint.TLSConfig.TLSCertPath = certPath
						keyPath, _ := handler.FileService.GetPathForTLSFile(folder, portainer.TLSFileKey)
						endpoint.TLSConfig.TLSKeyPath = keyPath
					} else {
						endpoint.TLSConfig.TLSCertPath = ""
						handler.FileService.DeleteTLSFile(folder, portainer.TLSFileCert)
						endpoint.TLSConfig.TLSKeyPath = ""
						handler.FileService.DeleteTLSFile(folder, portainer.TLSFileKey)
					}
				}

			} else {
				endpoint.TLSConfig.TLS = false
				endpoint.TLSConfig.TLSSkipVerify = false
				endpoint.TLSConfig.TLSCACertPath = ""
				endpoint.TLSConfig.TLSCertPath = ""
				endpoint.TLSConfig.TLSKeyPath = ""
				err = handler.FileService.DeleteTLSFiles(folder)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove TLS files from disk", err}
				}
			}

			if endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
				endpoint.TLSConfig.TLS = true
				endpoint.TLSConfig.TLSSkipVerify = true
			}
		}

		if (payload.URL != nil && *payload.URL != endpoint.URL) || (payload.TLS != nil && endpoint.TLSConfig.TLS != *payload.TLS) || endpoint.Type == portaineree.AzureEnvironment {
			handler.ProxyManager.DeleteEndpointProxy(endpoint.ID)
			_, err = handler.ProxyManager.CreateAndRegisterEndpointProxy(endpoint)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to register HTTP proxy for the environment", err}
			}
		}
	}

	if updateAuthorizations {
		if endpoint.Type == portaineree.KubernetesLocalEnvironment || endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
			err = handler.AuthorizationService.CleanNAPWithOverridePolicies(endpoint, nil)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update user authorizations", err}
			}
		}
	}

	if payload.ChangeWindow != nil {
		endpoint.ChangeWindow = *payload.ChangeWindow
	}

	err = handler.dataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist environment changes inside the database", err}
	}

	if updateAuthorizations {
		err = handler.AuthorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update user authorizations", err}
		}
	}

	if (endpointutils.IsEdgeEndpoint(endpoint)) && (groupIDChanged || tagsChanged) {
		relation, err := handler.dataStore.EndpointRelation().EndpointRelation(endpoint.ID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find environment relation inside the database", err}
		}

		endpointGroup, err := handler.dataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find environment group inside the database", err}
		}

		edgeGroups, err := handler.dataStore.EdgeGroup().EdgeGroups()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge groups from the database", err}
		}

		edgeStacks, err := handler.dataStore.EdgeStack().EdgeStacks()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge stacks from the database", err}
		}

		existingEdgeStacks := relation.EdgeStacks

		currentEdgeStackSet := map[portaineree.EdgeStackID]bool{}
		currentEndpointEdgeStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
		for _, edgeStackID := range currentEndpointEdgeStacks {
			currentEdgeStackSet[edgeStackID] = true
			if !existingEdgeStacks[edgeStackID] {
				err = handler.edgeService.AddStackCommand(endpoint, edgeStackID)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to store edge async command into the database", err}
				}
			}
		}

		for existingEdgeStackID := range existingEdgeStacks {
			if !currentEdgeStackSet[existingEdgeStackID] {
				err = handler.edgeService.RemoveStackCommand(endpoint.ID, existingEdgeStackID)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to store edge async command into the database", err}
				}
			}
		}

		relation.EdgeStacks = currentEdgeStackSet

		err = handler.dataStore.EndpointRelation().UpdateEndpointRelation(endpoint.ID, relation)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist environment relation changes inside the database", err}
		}
	}

	return response.JSON(w, endpoint)
}
