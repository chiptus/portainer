package deployments

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/datastore"
	"github.com/portainer/portainer-ee/api/internal/testhelpers"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const localhostCert = `-----BEGIN CERTIFICATE-----
MIIEOjCCAiKgAwIBAgIRALg8rJET2/9LjKSxHj0dQhYwDQYJKoZIhvcNAQELBQAw
FzEVMBMGA1UEAxMMUG9ydGFpbmVyIENBMB4XDTIzMTAxMTE5NDcxMVoXDTI1MDQx
MTE5NTM0MVowFDESMBAGA1UEAxMJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAx4nNGiwcCqUCxZyVLIHqvjTy20ZtZDVCedssTv1W5tmz
YqOIYGaW3CqzlRn6vBHu9bMHXef4+XfS0igKBn76MAKn5IcTccIWIal+5jq48pI3
c2FzQ3qNujX2zqZPjAjhJnVeVCP3kJu4wUtuubswLPBVLdktGa6EkL+8nu6o0Phw
6scV6s3gUmQk5/lpH4FIff8M7NAdTOxiFImQ1M0vplKtaEeiCnskpgyI8CbZl7X0
38Pu178W3+LqB7N4iMy2gKnBwjsXzw/+1dfUGkKjYdDBD+kNEKrQ4dwkjkrkQVdt
Z+GN26NvXHoeeyX/MLnVgdLbiIjvsf0DDIhabKqTcwIDAQABo4GDMIGAMA4GA1Ud
DwEB/wQEAwIDuDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0O
BBYEFPCefmK5Szzlfs8FRCa5+kRCIEWuMB8GA1UdIwQYMBaAFKZZ074SR/ajD3zE
gxpLGRvFT3XAMA8GA1UdEQQIMAaHBH8AAAEwDQYJKoZIhvcNAQELBQADggIBABcQ
/WPSUpuQvrcVBsmIlOMz74cDZYuIDls/mAcB/yP3mm+oWlO0qvH/F/BMs1P/bkgj
fByQZq8Zmi6/TEZNlGvW7KGx077VxDKi8jd1jL3gLDPmkFjYuGeIWQusgxBu1y3m
0WoTTqnkoism1mzV/dgNwrm3YQIV4H/fi9EEdQSm0UFRTKSAGBkwS7N2pmNb5yQO
U8glFpyznCv4evDJbs/JUUXKYExgFFhWUd25P7iBRLXg/BFfqdSTiUGUj/Msz0pO
Evqmq78eIiXjyyKSxzve6/mEIeq6AE3AC9zH+fwTd6Mhp+T2P/S/iO4EU19IMR4m
sbNBd6h/3GvRekO1KbqQ42awuMnxvWT0NVclSxiU1lMpAmRmk/w9z7wB3r4n7oh4
iiOTl5VSw1UBkcLDOJw+HB/FU2PdVFfIJKRfjLCZOGrcJX9vEcz7dYGpB5HrdqOc
/8q5j1g6f/pGE+20HITrtz6ChguETzqw5dLNeKeolC6bVH8yEtmpnP2n8VPnT9Di
V+hnONcJ+wd/dkBqabGr7LPG24Kj1F2Zp3CDDvJA94FaEsgaLfSg3JD+43uRCOWM
RuqU8bGuhQRqilR2dSIOrFaW2+MeUHsb24cUn/pkHqKpSg+RBEnf6QfGDlIgqYEl
19f/HFVBc/a8lM/D81lMyDbjQ9zH4LDYj4ipBbkL
-----END CERTIFICATE-----`

const localhostKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAx4nNGiwcCqUCxZyVLIHqvjTy20ZtZDVCedssTv1W5tmzYqOI
YGaW3CqzlRn6vBHu9bMHXef4+XfS0igKBn76MAKn5IcTccIWIal+5jq48pI3c2Fz
Q3qNujX2zqZPjAjhJnVeVCP3kJu4wUtuubswLPBVLdktGa6EkL+8nu6o0Phw6scV
6s3gUmQk5/lpH4FIff8M7NAdTOxiFImQ1M0vplKtaEeiCnskpgyI8CbZl7X038Pu
178W3+LqB7N4iMy2gKnBwjsXzw/+1dfUGkKjYdDBD+kNEKrQ4dwkjkrkQVdtZ+GN
26NvXHoeeyX/MLnVgdLbiIjvsf0DDIhabKqTcwIDAQABAoIBAQCqSP6BPG195A52
iEeCISksw9ERsou+fflKNvIcQvV7swP0xOyooERUhhiVwQMKpx9QDUXXLRV8CHch
JExR+OEYQdv4GhJM/b6XYafLYQfe80thKyQLzTXQWSdUeffe4OEMShODKOKoRUyp
oO9Qj9/wKfX3V6S2iwnU4dxdofztv+YP9rYQyjnhKbv/9OfeCp2Pb9eFKKRsA+QQ
xneDz1+wr8ToTuiTn8HBPNSeSAKvhzXuzyluI7VAetRloNgCtumrA9kpVbW2cDgE
Gk0q3RY125ejFELQO/cOJFuBsqoJlvPxzg8/vHyfyF9hFMqbqvcUw2e1eqHpnJd5
dP4+ZGYZAoGBAOOFuPXMLBts0rN9mfNbVfx36H+aOCL77SafZvWm0D+rH69QN3/q
/ZSWQEjwH5Tzn1e+NVcl/Um2vL/dIyEGBklXQ7yAyJo25gpEOD/rt1U94HKzMOwy
yKtsKghRAOx0piie7ORS6MGbEOQxU3/1Eg1uvd0qoSnALqJ/le75QpFXAoGBAOCD
aZQTszzDddr1cFPzLyqjIGJWfPcDYSONXVcCeQmhvC7mkfw9SWdIfku7JbdNgFYq
ZAAU0klsLX0lEe8f4A12FnHNylKoxmTWdE3wWPptejdA1KUgzt/2kNljgOMFuY0Q
rlCEW/Jabrg5aFMwVVG8bHLZR0xalfniDvXLvnFFAoGACdztJLKiIto31BIYz2Th
OF2WVZnA3ztej3MPioydsHThnb7zePcd4QgWZ1MJe3KIMMyNEWcTMNPcINEcSb0y
HpHK3OwURiMlG8LTUWoNe4OALFi6QTL+YfgBZnTkflucLFyfVlKFxobLV6kPvpdI
Hg7z6heD/wRWwTKYtFBX42cCgYBIeoQJ9rYlRqB0eEm0AEzYweLBfFRJVgD0/j0E
ytqSPnFG3s6AFLTur9t9zUPmwhFNP9Aaqp4cb9zbiq0YejzVe6rRQHMxbiTmBslz
I8VFyzPqRHahfE7sxGeMlm/UWlPFc34ipigcvA8EUBwaxv60LVUBWp2Gy7OhANZ9
iTHI1QKBgQCdHFj9dnbpaEHA426CoaPsyj5cv2nBLRf8p1cs71sq+qQOGlGJfajm
L9x22ol5c5rToZa1qKSnSdSDCud298MyRujMUy2UcUKHeNs3MK9AT41sDv266I7b
vJUUCFYm8+9p6gTVOcoMit+eGSwa81PCPEs1TnU1PV/PaDFeUhn/mg==
-----END RSA PRIVATE KEY-----`

type noopDeployer struct {
	SwarmStackDeployed      bool
	ComposeStackDeployed    bool
	KubernetesStackDeployed bool
}

// without unpacker
func (s *noopDeployer) DeploySwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	s.SwarmStackDeployed = true
	return nil
}

func (s *noopDeployer) DeployComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRecreate bool) error {
	s.ComposeStackDeployed = true
	return nil
}

func (s *noopDeployer) DeployKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User) error {
	s.KubernetesStackDeployed = true
	return nil
}

func (s *noopDeployer) RestartKubernetesStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, user *portaineree.User, resourceList []string) error {
	return nil
}

// with unpacker
func (s *noopDeployer) DeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, forcePullImage bool, forceRecreate bool) error {
	s.ComposeStackDeployed = true
	return nil
}

func (s *noopDeployer) UndeployRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	s.ComposeStackDeployed = true
	return nil
}

func (s *noopDeployer) StartRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry) error {
	return nil
}

func (s *noopDeployer) StopRemoteComposeStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	return nil
}

func (s *noopDeployer) DeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry, prune bool, pullImage bool) error {
	s.SwarmStackDeployed = true
	return nil
}

func (s *noopDeployer) UndeployRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	s.SwarmStackDeployed = true
	return nil
}

func (s *noopDeployer) StartRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint, registries []portaineree.Registry) error {
	return nil
}

func (s *noopDeployer) StopRemoteSwarmStack(stack *portaineree.Stack, endpoint *portaineree.Endpoint) error {
	return nil
}

func agentServer(t *testing.T) string {
	h := http.NewServeMux()

	h.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(portaineree.PortainerAgentHeader, "v2.19.0")
		w.Header().Set(portaineree.HTTPResponseAgentPlatform, strconv.Itoa(int(portaineree.AgentPlatformDocker)))

		response.Empty(w)
	})

	cert, err := tls.X509KeyPair([]byte(localhostCert), []byte(localhostKey))
	require.NoError(t, err)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	l, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	require.NoError(t, err)

	s := &http.Server{
		Handler: h,
	}

	go func() {
		err := s.Serve(l)
		require.ErrorIs(t, err, http.ErrServerClosed)
	}()

	t.Cleanup(func() {
		s.Shutdown(context.Background())
	})

	return "http://" + l.Addr().String()
}

func Test_redeployWhenChanged_FailsWhenCannotFindStack(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	err := RedeployWhenChanged(1, nil, store, nil, nil, nil)
	assert.Error(t, err)
	assert.Truef(t, strings.HasPrefix(err.Error(), "failed to get the stack"), "it isn't an error we expected: %v", err.Error())
}

func Test_redeployWhenChanged_DoesNothingWhenNotAGitBasedStack(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")
	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		URL: agentServer(t),
		TLSConfig: portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
	})
	assert.NoError(t, err, "error creating environment")
	stack := portaineree.Stack{ID: 1, CreatedBy: "admin"}
	err = store.Stack().Create(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}

	t.Run("can deploy docker compose stack", func(t *testing.T) {
		stack.Type = portaineree.DockerComposeStack
		store.Stack().Update(stack.ID, &stack)
		err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "oldHash"), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
	})
}

func Test_redeployWhenChanged_FailsWhenCannotClone(t *testing.T) {
	cloneErr := errors.New("failed to clone")
	_, store := datastore.MustNewTestStore(t, true, true)

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		URL: agentServer(t),
		TLSConfig: portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
	})
	assert.NoError(t, err, "error creating environment")

	err = store.Stack().Create(&portaineree.Stack{
		ID:        1,
		CreatedBy: "admin",
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		}})
	assert.NoError(t, err, "failed to create a test stack")

	err = RedeployWhenChanged(1, nil, store, testhelpers.NewGitService(cloneErr, "newHash"), nil, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cloneErr, "should failed to clone but didn't, check test setup")
}

func Test_redeployWhenChanged_ForceUpdateOn_WithAdditionalEnv(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	tmpDir := t.TempDir()

	err := store.Endpoint().Create(&portaineree.Endpoint{
		ID:  1,
		URL: agentServer(t),
		TLSConfig: portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
	})
	assert.NoError(t, err, "error creating environment")

	username := "user"
	err = store.User().Create(&portaineree.User{Username: username, Role: portaineree.AdministratorRole})
	assert.NoError(t, err, "error creating a user")

	stack := portaineree.Stack{
		ID:          1,
		EndpointID:  1,
		ProjectPath: tmpDir,
		UpdatedBy:   username,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portainer.AutoUpdateSettings{
			ForceUpdate: true,
		},
	}
	err = store.Stack().Create(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}

	options := &RedeployOptions{AdditionalEnvVars: []portainer.Pair{{Name: "version", Value: "latest"}}}

	t.Run("can deploy docker compose stack", func(t *testing.T) {
		stack.Type = portaineree.DockerComposeStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "oldHash"), nil, options)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
		result, _ := store.Stack().Read(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "oldHash")
	})

	t.Run("can deploy docker swarm stack", func(t *testing.T) {
		stack.Type = portaineree.DockerSwarmStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "oldHash"), nil, options)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.SwarmStackDeployed, true)
		result, _ := store.Stack().Read(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "oldHash")
	})

	t.Run("can deploy kube app", func(t *testing.T) {
		stack.Type = portaineree.KubernetesStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "oldHash"), nil, options)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.KubernetesStackDeployed, true)
		result, _ := store.Stack().Read(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "oldHash")
	})
}

func Test_redeployWhenChanged_RepoNotChanged_ForceUpdateOff(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	tmpDir := t.TempDir()

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		URL: agentServer(t),
		TLSConfig: portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
	})
	assert.NoError(t, err, "error creating environment")

	err = store.Stack().Create(&portaineree.Stack{
		ID:          1,
		CreatedBy:   "admin",
		ProjectPath: tmpDir,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portainer.AutoUpdateSettings{
			ForceUpdate: false,
		},
	})
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}
	err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "oldHash"), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, noopDeployer.ComposeStackDeployed, false)
	assert.Equal(t, noopDeployer.SwarmStackDeployed, false)
	assert.Equal(t, noopDeployer.KubernetesStackDeployed, false)
}

func Test_redeployWhenChanged_RepoNotChanged_ForceUpdateOff_ForcePullImageEnable(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	tmpDir := t.TempDir()

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	err = store.Endpoint().Create(&portaineree.Endpoint{
		ID: 0,
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled: false,
		},
		URL: agentServer(t),
		TLSConfig: portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
	})
	assert.NoError(t, err, "error creating environment")

	err = store.Stack().Create(&portaineree.Stack{
		ID:          1,
		CreatedBy:   "admin",
		ProjectPath: tmpDir,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portainer.AutoUpdateSettings{
			ForceUpdate:    false,
			ForcePullImage: true,
		},
		Type: portaineree.DockerComposeStack,
	})
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}
	err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "oldHash"), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
	assert.Equal(t, noopDeployer.SwarmStackDeployed, false)
	assert.Equal(t, noopDeployer.KubernetesStackDeployed, false)
}

func Test_redeployWhenChanged_RepoChanged_ForceUpdateOff(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	tmpDir := t.TempDir()

	err := store.Endpoint().Create(&portaineree.Endpoint{
		ID:  1,
		URL: agentServer(t),
		TLSConfig: portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
	})
	assert.NoError(t, err, "error creating environment")

	username := "user"
	err = store.User().Create(&portaineree.User{Username: username, Role: portaineree.AdministratorRole})
	assert.NoError(t, err, "error creating a user")

	stack := portaineree.Stack{
		ID:          1,
		EndpointID:  1,
		ProjectPath: tmpDir,
		UpdatedBy:   username,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portainer.AutoUpdateSettings{
			ForceUpdate: false,
		},
	}
	err = store.Stack().Create(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}

	t.Run("can deploy docker compose stack", func(t *testing.T) {
		stack.Type = portaineree.DockerComposeStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "newHash"), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.ComposeStackDeployed, true)
		result, _ := store.Stack().Read(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "newHash")
	})

	t.Run("can deploy docker swarm stack", func(t *testing.T) {
		stack.Type = portaineree.DockerSwarmStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "newHash"), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.SwarmStackDeployed, true)
		result, _ := store.Stack().Read(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "newHash")
	})

	t.Run("can deploy kube app", func(t *testing.T) {
		stack.Type = portaineree.KubernetesStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "newHash"), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, noopDeployer.KubernetesStackDeployed, true)
		result, _ := store.Stack().Read(stack.ID)
		assert.Equal(t, result.GitConfig.ConfigHash, "newHash")
	})
}

func Test_getUserRegistries(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	endpointID := 123

	admin := &portaineree.User{ID: 1, Username: "admin", Role: portaineree.AdministratorRole}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	user := &portaineree.User{ID: 2, Username: "user", Role: portaineree.StandardUserRole}
	err = store.User().Create(user)
	assert.NoError(t, err, "error creating a user")

	team := portainer.Team{ID: 1, Name: "team"}

	store.TeamMembership().Create(&portainer.TeamMembership{
		ID:     1,
		UserID: user.ID,
		TeamID: team.ID,
		Role:   portaineree.TeamMember,
	})

	registryReachableByUser := portaineree.Registry{
		ID: 1,
		RegistryAccesses: portainer.RegistryAccesses{
			portainer.EndpointID(endpointID): {
				UserAccessPolicies: map[portainer.UserID]portainer.AccessPolicy{
					user.ID: {RoleID: portaineree.RoleIDStandardUser},
				},
			},
		},
	}
	err = store.Registry().Create(&registryReachableByUser)
	assert.NoError(t, err, "couldn't create a registry")

	registryReachableByTeam := portaineree.Registry{
		ID: 2,
		RegistryAccesses: portainer.RegistryAccesses{
			portainer.EndpointID(endpointID): {
				TeamAccessPolicies: map[portainer.TeamID]portainer.AccessPolicy{
					team.ID: {RoleID: portaineree.RoleIDStandardUser},
				},
			},
		},
	}
	err = store.Registry().Create(&registryReachableByTeam)
	assert.NoError(t, err, "couldn't create a registry")

	registryRestricted := portaineree.Registry{
		ID: 3,
		RegistryAccesses: portainer.RegistryAccesses{
			portainer.EndpointID(endpointID): {
				UserAccessPolicies: map[portainer.UserID]portainer.AccessPolicy{
					user.ID + 100: {RoleID: portaineree.RoleIDStandardUser},
				},
			},
		},
	}
	err = store.Registry().Create(&registryRestricted)
	assert.NoError(t, err, "couldn't create a registry")

	t.Run("admin should has access to all registries", func(t *testing.T) {
		registries, err := getUserRegistries(store, admin, portainer.EndpointID(endpointID))
		assert.NoError(t, err)
		assert.ElementsMatch(t, []portaineree.Registry{registryReachableByUser, registryReachableByTeam, registryRestricted}, registries)
	})

	t.Run("regular user has access to registries allowed to him and/or his team", func(t *testing.T) {
		registries, err := getUserRegistries(store, user, portainer.EndpointID(endpointID))
		assert.NoError(t, err)
		assert.ElementsMatch(t, []portaineree.Registry{registryReachableByUser, registryReachableByTeam}, registries)
	})
}

func newChangeWindow(start, end string) *portaineree.Endpoint {
	return &portaineree.Endpoint{
		ChangeWindow: portaineree.EndpointChangeWindow{
			Enabled:   true,
			StartTime: start,
			EndTime:   end,
		},
	}
}

type MockClock struct {
	timeNow string
}

func (mc MockClock) Now() time.Time {
	gt, _ := time.Parse("15:04", mc.timeNow)
	return gt
}

func NewMockClock(t string) Clock {
	return MockClock{
		timeNow: t,
	}
}

func Test_updateAllowed(t *testing.T) {
	is := assert.New(t)

	// Ensure change window works to the following rules including edge cases (start <= t < endtime) = true
	t.Run("updateAllowed is true inside the time window", func(t *testing.T) {
		endpoint := newChangeWindow("22:00", "23:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("22:30"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed is true when time equal to start time", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "23:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("10:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed is false when time equal to end time (exclusive)", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "11:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("11:00"))
		is.NoError(err)
		is.Equal(false, allowed, "updateAllowed should be false")
	})

	t.Run("updateAllowed is true when start and end time are equal", func(t *testing.T) {
		// we treat this as 24hrs window which means fully on
		endpoint := newChangeWindow("10:00", "10:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("12:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed is false when time outside the time window", func(t *testing.T) {
		endpoint := newChangeWindow("00:00", "05:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("07:00"))
		is.NoError(err)
		is.Equal(false, allowed, "updateAllowed should be false")
	})

	t.Run("updateAllowed when end time spans over to the next day", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "02:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("11:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day and time within and next day", func(t *testing.T) {
		endpoint := newChangeWindow("10:00", "02:00")
		allowed, err := updateAllowed(endpoint, NewMockClock("01:00"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day and time within current day", func(t *testing.T) {
		endpoint := newChangeWindow("10:35", "02:45")
		allowed, err := updateAllowed(endpoint, NewMockClock("10:47"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day and time within next day same hour", func(t *testing.T) {
		endpoint := newChangeWindow("10:30", "02:27")
		allowed, err := updateAllowed(endpoint, NewMockClock("02:20"))
		is.NoError(err)
		is.Equal(true, allowed, "updateAllowed should be true")
	})

	t.Run("updateAllowed when end time spans over to the next day time outside window next day same hour", func(t *testing.T) {
		endpoint := newChangeWindow("10:35", "02:45")
		allowed, err := updateAllowed(endpoint, NewMockClock("02:47"))
		is.NoError(err)
		is.Equal(false, allowed, "updateAllowed should be false")
	})
}

type mockGitService struct {
	id string
}

func newMockGitService(id string) portainer.GitService {
	return &mockGitService{
		id: id,
	}
}

func (g *mockGitService) CloneRepository(destination, repositoryURL, referenceName, username, password string, tlsSkipVerify bool) error {
	return os.Mkdir(destination, 0644)
}

func (g *mockGitService) LatestCommitID(repositoryURL, referenceName, username, password string, tlsSkipVerify bool) (string, error) {
	return g.id, nil
}

func (g *mockGitService) ListRefs(repositoryURL, username, password string, hardRefresh bool, tlsSkipVerify bool) ([]string, error) {
	return nil, nil
}

func (g *mockGitService) ListFiles(repositoryURL, referenceName, username, password string, dirOnly, hardRefresh bool, includedExts []string, tlsSkipVerify bool) ([]string, error) {
	return nil, nil
}

func Test_redeployWhenChanged_RepoChanged_VersionFolderRemoved(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	tmpDir := t.TempDir()

	err := store.Endpoint().Create(&portaineree.Endpoint{
		ID:  1,
		URL: agentServer(t),
		TLSConfig: portainer.TLSConfiguration{
			TLS:           true,
			TLSSkipVerify: true,
		},
	})
	assert.NoError(t, err, "error creating environment")

	username := "user"
	err = store.User().Create(&portaineree.User{Username: username, Role: portaineree.AdministratorRole})
	assert.NoError(t, err, "error creating a user")

	stack := portaineree.Stack{
		ID:          1,
		EndpointID:  1,
		ProjectPath: tmpDir,
		UpdatedBy:   username,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portainer.AutoUpdateSettings{
			ForceUpdate: false,
		},
	}
	err = store.Stack().Create(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	oldVersionFolder := filesystem.JoinPaths(stack.ProjectPath, stack.GitConfig.ConfigHash)
	err = os.MkdirAll(oldVersionFolder, 0644)
	assert.NoError(t, err, "failed to create a test stack version folder")

	noopDeployer := &noopDeployer{}

	t.Run("can remove old version folder", func(t *testing.T) {
		stack.Type = portaineree.DockerComposeStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, newMockGitService("newHash"), nil, nil)
		assert.NoError(t, err)
		assert.NoDirExists(t, oldVersionFolder, "old version folder should be removed")
		newVersionFolder := filesystem.JoinPaths(stack.ProjectPath, "newHash")
		assert.DirExists(t, newVersionFolder, "new version folder should be created")
		stack.GitConfig.ConfigHash = "newHash"
	})

	t.Run("can remove old version folder after another deployment", func(t *testing.T) {
		stack.Type = portaineree.DockerSwarmStack
		store.Stack().Update(stack.ID, &stack)

		err = RedeployWhenChanged(1, noopDeployer, store, newMockGitService("secondHash"), nil, nil)
		assert.NoError(t, err)
		oldVersionFolder := filesystem.JoinPaths(stack.ProjectPath, "newHash")
		assert.NoDirExists(t, oldVersionFolder, "old version folder should be removed")
		newVersionFolder := filesystem.JoinPaths(stack.ProjectPath, "secondHash")
		assert.DirExists(t, newVersionFolder, "new version folder should be created")
	})
}

func Test_redeployWhenChanged_NoDeployWhenEnvironmentOffline(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	tmpDir := t.TempDir()

	admin := &portaineree.User{ID: 1, Username: "admin"}
	err := store.User().Create(admin)
	assert.NoError(t, err, "error creating an admin")

	err = store.Endpoint().Create(&portaineree.Endpoint{ID: 0})
	assert.NoError(t, err, "error creating environment")

	err = store.Stack().Create(&portaineree.Stack{
		ID:          1,
		CreatedBy:   "admin",
		ProjectPath: tmpDir,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		},
		AutoUpdate: &portainer.AutoUpdateSettings{
			ForceUpdate:    false,
			ForcePullImage: true,
		},
		Type: portaineree.DockerComposeStack,
	})
	assert.NoError(t, err, "failed to create a test stack")

	noopDeployer := &noopDeployer{}
	err = RedeployWhenChanged(1, noopDeployer, store, testhelpers.NewGitService(nil, "oldHash"), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, noopDeployer.ComposeStackDeployed, false)
	assert.Equal(t, noopDeployer.SwarmStackDeployed, false)
	assert.Equal(t, noopDeployer.KubernetesStackDeployed, false)
}
