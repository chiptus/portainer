package edgestacks

import (
	"os"
	"path/filepath"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_convertAndStoreKubeManifestIfNeeded(t *testing.T) {
	composeFile := `
		version: "3"

		services:
		  balancer:
				image: nginx
	`
	composeFileName := "docker-compose.yml"

	// kompose := ``

	projectPath := t.TempDir()
	f, _ := os.Create(filepath.Join(projectPath, composeFileName))
	f.WriteString(composeFile)

	edgeStack := &portainer.EdgeStack{
		ProjectPath: projectPath,
		EntryPoint:  "docker-compose.yml",
	}

	relatedIds := []portainer.EndpointID{1}

	endpoint := &portainer.Endpoint{
		ID:   relatedIds[0],
		Type: portainer.EdgeAgentOnKubernetesEnvironment,
	}

	dataStore := testhelpers.NewDatastore(testhelpers.WithEndpoints([]portainer.Endpoint{*endpoint}))

	h := NewHandler(nil)
	h.DataStore = dataStore
	err := h.convertAndStoreKubeManifestIfNeeded(edgeStack, relatedIds)

	assert.NoError(t, err)

}
