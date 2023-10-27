package license

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/portainer/liblicense/v3"
	portaineree "github.com/portainer/portainer-ee/api"
	statusutil "github.com/portainer/portainer-ee/api/internal/nodes"
	portainer "github.com/portainer/portainer/api"
	"github.com/rs/zerolog/log"
)

const (
	syncInterval = time.Hour * 24

	// a period of time after which license overuse restrictions will be enforced.
	// default value: 10 days
	overuseGracePeriodInSeconds = 10 * 24 * 60 * 60
)

func (service *Service) startSyncLoop() error {
	err := service.SyncLicenses()
	if err != nil {
		log.Err(err).Msg("failed initial license sync")
	}

	ticker := time.NewTicker(syncInterval)

	go func() {
		for {
			select {
			case <-service.shutdownCtx.Done():
				log.Debug().Msg("shutting down License service")
				ticker.Stop()

				return
			case <-ticker.C:
				service.SyncLicenses()
			}
		}
	}()

	return nil
}

// syncLiceses checks all of the licenses with the license server to determine
// if any of them have been revoked or otherwise to not exist on the server.
func (service *Service) SyncLicenses() error {
	licenses, err := service.dataStore.License().Licenses()
	if err != nil {
		return err
	}

	var validLicenses []liblicense.PortainerLicense
	for _, l := range licenses {
		l := ParseLicense(l.LicenseKey, service.expireAbsolute, l.Revoked)
		if !l.Revoked {
			validLicenses = append(validLicenses, l)
		}
	}

	meta, err := service.Metadata(licenses)
	if err != nil {
		return err
	}
	info := liblicense.CheckInInfo{
		Licenses: validLicenses,
		Meta:     meta,
	}
	resp, err := liblicense.CheckIn(info)
	if err != nil {
		log.Err(err).Msg("invalid license found")
	}

	for _, l := range validLicenses {
		if !resp[l.LicenseKey] {
			service.revokeLicense(l.LicenseKey)
		}
	}

	licenses, err = service.dataStore.License().Licenses()
	if err != nil {
		return err
	}
	service.licenses = licenses
	return nil
}

// Metadata gathers basic metadata about the installation.
func (service *Service) Metadata(licenses []liblicense.PortainerLicense) (liblicense.Metadata, error) {
	var meta liblicense.Metadata

	// InstanceID should be unique per install.
	instanceID, err := service.dataStore.Version().InstanceID()
	if err != nil {
		log.Err(err).Msg("failed fetching instanceID")
		return meta, err
	}

	var licenseKeys []string
	for _, l := range licenses {
		licenseKeys = append(licenseKeys, l.LicenseKey)
	}

	// Gather environments, roles and users.
	// endpointMap is a mapping of endpoint types to a count of their usage.
	// We don't use portainer.EndpointType as we need a bit more granularity
	// than is offered from that internal type.
	// 1 = LocalDocker
	// 2 = APIDocker
	// 3 = AgentDocker
	// 4 = LocalAzure
	// 5 = EdgeDocker
	// 6 = EdgeAsyncDocker
	// 7 = LocalKubernetes
	// 8 = AgentKubernetes
	// 9 = EdgeKubernetes
	// 10 = EdgeAsyncKubernetes
	// 11 = KubeConfigKubernetes
	// 12 = Microk8sKubernetes
	// 13 = CloudCivoKubernetes
	// 14 = CloudDigitalOceanKubernetes
	// 15 = CloudLinodeKubernetes
	// 16 = CloudGKEKubernetes
	// 17 = CloudAzureKubernetes
	// 18 = CloudAmazonKubernetes
	// roleMap is a map of each RoleID to UserIDs assigned to it. You can use
	// this to determine how many users are assigned a particular role or if
	// users are assigned different roles in different endpoints.
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		log.Err(err).Msg("failed fetching endpoints")
		return meta, err
	}

	roleMap := make(map[portainer.RoleID]map[portainer.UserID]struct{})
	endpointMap := make(map[string]int)
	for _, e := range endpoints {
		for userID, accessPolicy := range e.UserAccessPolicies {
			if roleMap[accessPolicy.RoleID] == nil {
				roleMap[accessPolicy.RoleID] = make(map[portainer.UserID]struct{})
			}
			roleMap[accessPolicy.RoleID][userID] = struct{}{}
		}

		eType := endpointMetadataType(e)
		endpointMap[instanceID+":"+strconv.Itoa(eType)] += 1
	}

	var roles []liblicense.MetadataRole
	for r, u := range roleMap {
		roles = append(roles, liblicense.MetadataRole{
			ID:         instanceID + ":" + strconv.Itoa(int(r)),
			InstanceID: instanceID,
			Role:       int(r),
			Users:      len(u),
		})
	}

	var endpointTypes []liblicense.MetadataEndpointType
	for e, c := range endpointMap {
		instanceID, eType, ok := strings.Cut(e, ":")
		if !ok {
			continue
		}
		t, err := strconv.Atoi(eType)
		if err != nil {
			continue
		}
		endpointTypes = append(endpointTypes, liblicense.MetadataEndpointType{
			ID:         e,
			InstanceID: instanceID,
			Type:       t,
			Count:      c,
		})
	}

	var adminUsers, nonAdminUsers int
	users, err := service.dataStore.User().ReadAll()
	for _, user := range users {
		if user.Role == 1 {
			adminUsers += 1
		} else {
			nonAdminUsers += 1
		}
	}
	if err != nil {
		log.Err(err).Msg("failed fetching users")
		return meta, err
	}

	meta.InstanceID = instanceID
	meta.AuthMethod = service.authMethod()
	meta.LicenseKeys = licenseKeys
	meta.PortainerVersion = portaineree.APIVersion
	meta.CPUCount = service.cpuCount()
	meta.NodesLicensed = service.Info().Nodes
	meta.NodesUsed = statusutil.NodesCount(endpoints)
	meta.AdminUsers = adminUsers
	meta.NonAdminUsers = nonAdminUsers
	meta.Roles = roles
	meta.EndpointTypes = endpointTypes
	return meta, nil
}

type MetadataAuthMethod int

const (
	MetadataAuthUnknown MetadataAuthMethod = iota
	MetadataAuthInternal
	MetadataAuthLDAPCustom
	MetadataAuthOauthCustom
	MetadataAuthLDAPOpenLDAP
	MetadataAuthLDAPAD
	MetadataAuthOauthMicrosoft
	MetadataAuthOauthGoogle
	MetadataAuthOauthGithub
)

func (service *Service) authMethod() int {
	settings, err := service.dataStore.Settings().Settings()
	if err != nil {
		log.Err(err).Msg("failed fetching settings")
		return 0
	}
	switch settings.AuthenticationMethod {
	case portainer.AuthenticationInternal:
		return int(MetadataAuthInternal)
	case portainer.AuthenticationLDAP:
		switch settings.LDAPSettings.ServerType {
		case portaineree.LDAPServerOpenLDAP:
			return int(MetadataAuthLDAPOpenLDAP)
		case portaineree.LDAPServerAD:
			return int(MetadataAuthLDAPAD)
		default:
			return int(MetadataAuthLDAPCustom)
		}
	case portainer.AuthenticationOAuth:
		switch settings.OAuthSettings.AuthorizationURI {
		case "https://login.microsoftonline.com/TENANT_ID/oauth2/v2.0/authorize":
			return int(MetadataAuthOauthMicrosoft)
		case "https://accounts.google.com/o/oauth2/auth":
			return int(MetadataAuthOauthGoogle)
		case "https://github.com/login/oauth/authorize":
			return int(MetadataAuthOauthGithub)
		default:
			return int(MetadataAuthOauthCustom)
		}
	default:
		return int(MetadataAuthUnknown)
	}
}

func (service *Service) cpuCount() int {
	var count int
	endpoints, err := service.dataStore.Endpoint().Endpoints()
	if err != nil {
		log.Err(err).Msg("failed fetching endpoints")
		return count
	}

	for _, e := range endpoints {
		if e.Status != portaineree.EndpointStatusUp {
			continue
		}
		switch e.Type {
		case portaineree.DockerEnvironment, portaineree.AgentOnDockerEnvironment:
			dcl, err := service.dockerClientFactory.CreateClient(&e, "", nil)
			if err != nil {
				log.Err(err).Msg("failed fetching dockerclient for cpu count")
				continue
			}
			info, err := dcl.Info(context.Background())
			if err != nil {
				log.Err(err).Msg("failed reading cpu count")
				continue
			}
			count += info.NCPU
		case portaineree.KubernetesLocalEnvironment, portaineree.AgentOnKubernetesEnvironment:
			kcl, err := service.kubernetesClientFactory.GetKubeClient(&e)
			if err != nil {
				log.Err(err).Msg("failed fetching kubeclient for cpu count")
				continue
			}
			c, err := kcl.GetCPUCount()
			if err != nil {
				log.Err(err).Msg("failed reading cpu count")
				continue
			}
			count += c
		}
	}

	return count
}

type MetadataEndpointType int

const (
	_ MetadataEndpointType = iota
	MetadataLocalDocker
	MetadataAPIDocker
	MetadataAgentDocker
	MetadataLocalAzure
	MetadataEdgeDocker
	MetadataEdgeAsyncDocker
	MetadataLocalKubernetes
	MetadataAgentKubernetes
	MetadataEdgeKubernetes
	MetadataEdgeAsyncKubernetes
	MetadataKubeConfigKubernetes
	MetadataMicrok8sKubernetes
	MetadataCloudCivoKubernetes
	MetadataCloudDigitalOceanKubernetes
	MetadataCloudLinodeKubernetes
	MetadataCloudGKEKubernetes
	MetadataCloudAzureKubernetes
	MetadataCloudAmazonKubernetes
)

func endpointMetadataType(e portaineree.Endpoint) int {
	var t int
	switch e.Type {
	case portainer.DockerEnvironment:
		t = int(MetadataAPIDocker)
		if strings.HasPrefix(e.URL, "unix://") {
			t = int(MetadataLocalDocker)
		}
	case portainer.AgentOnDockerEnvironment:
		t = int(MetadataAgentDocker)
	case portainer.AzureEnvironment:
		t = int(MetadataLocalAzure)
	case portainer.EdgeAgentOnDockerEnvironment:
		t = int(MetadataEdgeDocker)
		if e.Edge.AsyncMode {
			t = int(MetadataEdgeAsyncDocker)
		}
	case portainer.KubernetesLocalEnvironment:
		t = int(MetadataLocalKubernetes)
	case portainer.AgentOnKubernetesEnvironment:
		t = int(MetadataAgentKubernetes)
		if e.CloudProvider != nil {
			switch e.CloudProvider.Provider {
			case portaineree.CloudProviderCivo:
				t = int(MetadataCloudCivoKubernetes)
			case portaineree.CloudProviderDigitalOcean:
				t = int(MetadataCloudDigitalOceanKubernetes)
			case portaineree.CloudProviderLinode:
				t = int(MetadataCloudLinodeKubernetes)
			case portaineree.CloudProviderGKE:
				t = int(MetadataCloudGKEKubernetes)
			case portaineree.CloudProviderAzure:
				t = int(MetadataCloudAzureKubernetes)
			case portaineree.CloudProviderAmazon:
				t = int(MetadataCloudAmazonKubernetes)
			case portaineree.CloudProviderMicrok8s:
				t = int(MetadataMicrok8sKubernetes)
			case portaineree.CloudProviderKubeConfig:
				t = int(MetadataKubeConfigKubernetes)
			}
		}
	case portainer.EdgeAgentOnKubernetesEnvironment:
		t = int(MetadataEdgeKubernetes)
		if e.Edge.AsyncMode {
			t = int(MetadataEdgeAsyncKubernetes)
		}
	}
	return t
}

func RecalculateLicenseUsage(licenseService portaineree.LicenseService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(rw, r)

		if licenseService != nil {
			licenseService.SyncLicenses()
		}
	})
}
