package handler

import (
	"net/http"
	"strings"

	"github.com/portainer/portainer-ee/api/http/handler/auth"
	"github.com/portainer/portainer-ee/api/http/handler/backup"
	"github.com/portainer/portainer-ee/api/http/handler/cloudcredentials"
	"github.com/portainer/portainer-ee/api/http/handler/customtemplates"
	"github.com/portainer/portainer-ee/api/http/handler/docker"
	"github.com/portainer/portainer-ee/api/http/handler/edgegroups"
	"github.com/portainer/portainer-ee/api/http/handler/edgejobs"
	"github.com/portainer/portainer-ee/api/http/handler/edgestacks"
	"github.com/portainer/portainer-ee/api/http/handler/edgetemplates"
	"github.com/portainer/portainer-ee/api/http/handler/edgeupdateschedules"
	"github.com/portainer/portainer-ee/api/http/handler/endpointedge"
	"github.com/portainer/portainer-ee/api/http/handler/endpointgroups"
	"github.com/portainer/portainer-ee/api/http/handler/endpointproxy"
	"github.com/portainer/portainer-ee/api/http/handler/endpoints"
	"github.com/portainer/portainer-ee/api/http/handler/file"
	"github.com/portainer/portainer-ee/api/http/handler/gitops"
	"github.com/portainer/portainer-ee/api/http/handler/helm"
	"github.com/portainer/portainer-ee/api/http/handler/hostmanagement/fdo"
	"github.com/portainer/portainer-ee/api/http/handler/hostmanagement/openamt"
	"github.com/portainer/portainer-ee/api/http/handler/kaas"
	"github.com/portainer/portainer-ee/api/http/handler/kubernetes"
	"github.com/portainer/portainer-ee/api/http/handler/ldap"
	"github.com/portainer/portainer-ee/api/http/handler/licenses"
	"github.com/portainer/portainer-ee/api/http/handler/motd"
	"github.com/portainer/portainer-ee/api/http/handler/nomad"
	"github.com/portainer/portainer-ee/api/http/handler/registries"
	"github.com/portainer/portainer-ee/api/http/handler/resourcecontrols"
	"github.com/portainer/portainer-ee/api/http/handler/roles"
	"github.com/portainer/portainer-ee/api/http/handler/settings"
	"github.com/portainer/portainer-ee/api/http/handler/sshkey"
	"github.com/portainer/portainer-ee/api/http/handler/ssl"
	"github.com/portainer/portainer-ee/api/http/handler/stacks"
	"github.com/portainer/portainer-ee/api/http/handler/storybook"
	"github.com/portainer/portainer-ee/api/http/handler/system"
	"github.com/portainer/portainer-ee/api/http/handler/tags"
	"github.com/portainer/portainer-ee/api/http/handler/teammemberships"
	"github.com/portainer/portainer-ee/api/http/handler/teams"
	"github.com/portainer/portainer-ee/api/http/handler/templates"
	"github.com/portainer/portainer-ee/api/http/handler/upload"
	"github.com/portainer/portainer-ee/api/http/handler/useractivity"
	"github.com/portainer/portainer-ee/api/http/handler/users"
	"github.com/portainer/portainer-ee/api/http/handler/webhooks"
	"github.com/portainer/portainer-ee/api/http/handler/websocket"
)

// Handler is a collection of all the service handlers.
type Handler struct {
	AuthHandler               *auth.Handler
	BackupHandler             *backup.Handler
	CustomTemplatesHandler    *customtemplates.Handler
	DockerHandler             *docker.Handler
	EdgeGroupsHandler         *edgegroups.Handler
	EdgeJobsHandler           *edgejobs.Handler
	EdgeUpdateScheduleHandler *edgeupdateschedules.Handler
	EdgeStacksHandler         *edgestacks.Handler
	EdgeTemplatesHandler      *edgetemplates.Handler
	EndpointEdgeHandler       *endpointedge.Handler
	EndpointGroupHandler      *endpointgroups.Handler
	EndpointHandler           *endpoints.Handler
	EndpointHelmHandler       *helm.Handler
	EndpointProxyHandler      *endpointproxy.Handler
	GitOperationHandler       *gitops.Handler
	HelmTemplatesHandler      *helm.Handler
	KaasHandler               *kaas.Handler
	KubernetesHandler         *kubernetes.Handler
	FileHandler               *file.Handler
	LDAPHandler               *ldap.Handler
	MOTDHandler               *motd.Handler
	LicenseHandler            *licenses.Handler
	RegistryHandler           *registries.Handler
	ResourceControlHandler    *resourcecontrols.Handler
	RoleHandler               *roles.Handler
	SettingsHandler           *settings.Handler
	SSLHandler                *ssl.Handler
	OpenAMTHandler            *openamt.Handler
	FDOHandler                *fdo.Handler
	StackHandler              *stacks.Handler
	SystemHandler             *system.Handler
	StorybookHandler          *storybook.Handler
	TagHandler                *tags.Handler
	TeamMembershipHandler     *teammemberships.Handler
	TeamHandler               *teams.Handler
	TemplatesHandler          *templates.Handler
	UploadHandler             *upload.Handler
	UserHandler               *users.Handler
	UserActivityHandler       *useractivity.Handler
	WebSocketHandler          *websocket.Handler
	WebhookHandler            *webhooks.Handler
	NomadHandler              *nomad.Handler
	CloudCredentialsHandler   *cloudcredentials.Handler
	SSHKeyHandler             *sshkey.Handler
}

// @title PortainerEE API
// @version 2.18.0
// @description.markdown api-description.md
// @termsOfService

// @contact.email info@portainer.io

// @host
// @BasePath /api
// @schemes http https

// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securitydefinitions.apikey jwt
// @in header
// @name Authorization

// @tag.name auth
// @tag.description Authenticate against Portainer HTTP API
// @tag.name custom_templates
// @tag.description Manage Custom Templates
// @tag.name dockerhub
// @tag.description Manage how Portainer connects to the DockerHub
// @tag.name edge_groups
// @tag.description Manage Edge Groups
// @tag.name edge_jobs
// @tag.description Manage Edge Jobs
// @tag.name edge_stacks
// @tag.description Manage Edge Stacks
// @tag.name edge_templates
// @tag.description Manage Edge Templates
// @tag.name edge
// @tag.description Manage Edge related environment(endpoint) settings
// @tag.name endpoints
// @tag.description Manage Docker environments(endpoints)
// @tag.name endpoint_groups
// @tag.description Manage environment(endpoint) groups
// @tag.name gitops
// @tag.description Operate git repository
// @tag.name kubernetes
// @tag.description Manage Kubernetes cluster
// @tag.name motd
// @tag.description Fetch the message of the day
// @tag.name registries
// @tag.description Manage Docker registries
// @tag.name resource_controls
// @tag.description Manage access control on Docker resources
// @tag.name roles
// @tag.description Manage roles
// @tag.name settings
// @tag.description Manage Portainer settings
// @tag.name users
// @tag.description Manage users
// @tag.name tags
// @tag.description Manage tags
// @tag.name teams
// @tag.description Manage teams
// @tag.name team_memberships
// @tag.description Manage team memberships
// @tag.name templates
// @tag.description Manage App Templates
// @tag.name stacks
// @tag.description Manage stacks
// @tag.name upload
// @tag.description Upload files
// @tag.name webhooks
// @tag.description Manage webhooks
// @tag.name websocket
// @tag.description Create exec sessions using websockets
// @tag.name status
// @tag.description Information about the Portainer instance
// @tag.name system
// @tag.description Manage Portainer system

// ServeHTTP delegates a request to the appropriate subhandler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/endpoints") && strings.Contains(r.URL.Path, "/edge/"):
		h.EndpointEdgeHandler.ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/auth"):
		http.StripPrefix("/api", h.AuthHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/backup"):
		http.StripPrefix("/api", h.BackupHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/restore"):
		http.StripPrefix("/api", h.BackupHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/custom_templates"):
		http.StripPrefix("/api", h.CustomTemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_update_schedules"):
		http.StripPrefix("/api", h.EdgeUpdateScheduleHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_stacks"):
		http.StripPrefix("/api", h.EdgeStacksHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_groups"):
		http.StripPrefix("/api", h.EdgeGroupsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_jobs"):
		http.StripPrefix("/api", h.EdgeJobsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_templates"):
		http.StripPrefix("/api", h.EdgeTemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/endpoint_groups"):
		http.StripPrefix("/api", h.EndpointGroupHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/cloudcredentials"):
		http.StripPrefix("/api", h.CloudCredentialsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/cloud"):
		http.StripPrefix("/api", h.KaasHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/sshkeygen"):
		http.StripPrefix("/api", h.SSHKeyHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/kubernetes"):
		http.StripPrefix("/api", h.KubernetesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/docker"):
		http.StripPrefix("/api/docker", h.DockerHandler).ServeHTTP(w, r)

	// Helm subpath under kubernetes -> /api/endpoints/{id}/kubernetes/helm
	case strings.HasPrefix(r.URL.Path, "/api/endpoints/") && strings.Contains(r.URL.Path, "/kubernetes/helm"):
		http.StripPrefix("/api/endpoints", h.EndpointHelmHandler).ServeHTTP(w, r)

	case strings.HasPrefix(r.URL.Path, "/api/endpoints"):
		switch {
		case strings.Contains(r.URL.Path, "/docker/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/kubernetes/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/azure/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/agent/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		default:
			http.StripPrefix("/api", h.EndpointHandler).ServeHTTP(w, r)
		}
	case strings.HasPrefix(r.URL.Path, "/api/gitops"):
		http.StripPrefix("/api", h.GitOperationHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/ldap"):
		http.StripPrefix("/api", h.LDAPHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/licenses"):
		http.StripPrefix("/api", h.LicenseHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/motd"):
		http.StripPrefix("/api", h.MOTDHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/registries"):
		http.StripPrefix("/api", h.RegistryHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/resource_controls"):
		http.StripPrefix("/api", h.ResourceControlHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/roles"):
		http.StripPrefix("/api", h.RoleHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/settings"):
		http.StripPrefix("/api", h.SettingsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/stacks"):
		http.StripPrefix("/api", h.StackHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/status"):
		http.StripPrefix("/api", h.SystemHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/system"):
		http.StripPrefix("/api", h.SystemHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/tags"):
		http.StripPrefix("/api", h.TagHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/templates/helm"):
		http.StripPrefix("/api", h.HelmTemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/templates"):
		http.StripPrefix("/api", h.TemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/upload"):
		http.StripPrefix("/api", h.UploadHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/users"):
		http.StripPrefix("/api", h.UserHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/useractivity"):
		http.StripPrefix("/api", h.UserActivityHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/ssl"):
		http.StripPrefix("/api", h.SSLHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/open_amt"):
		http.StripPrefix("/api", h.OpenAMTHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/fdo"):
		http.StripPrefix("/api", h.FDOHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/teams"):
		http.StripPrefix("/api", h.TeamHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/team_memberships"):
		http.StripPrefix("/api", h.TeamMembershipHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/websocket"):
		http.StripPrefix("/api", h.WebSocketHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/webhooks"):
		http.StripPrefix("/api", h.WebhookHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/storybook"):
		http.StripPrefix("/storybook", h.StorybookHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/nomad"):
		http.StripPrefix("/api", h.NomadHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/"):
		h.FileHandler.ServeHTTP(w, r)
	}
}
