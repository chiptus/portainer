package stackutils

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/kubernetes"
	k "github.com/portainer/portainer-ee/api/kubernetes"
	"github.com/portainer/portainer/api/filesystem"
)

// ResourceControlID returns the stack resource control id
func ResourceControlID(endpointID portaineree.EndpointID, name string) string {
	return fmt.Sprintf("%d_%s", endpointID, name)
}

// GetStackFilePaths returns a list of file paths based on stack project path
func GetStackFilePaths(stack *portaineree.Stack) []string {
	var filePaths []string
	for _, file := range append([]string{stack.EntryPoint}, stack.AdditionalFiles...) {
		filePaths = append(filePaths, filesystem.JoinPaths(stack.ProjectPath, file))
	}
	return filePaths
}

type DeploymentFiles struct {
	FilePaths  []string
	Namespaces []string
}

// CreateTempK8SDeploymentFiles reads manifest files from original stack project path
// then add app labels into the file contents and create temp files for deployment
// return temp file paths and temp dir
func CreateTempK8SDeploymentFiles(stack *portaineree.Stack, kubeDeployer portaineree.KubernetesDeployer, appLabels k.KubeAppLabels) (*DeploymentFiles, func(), error) {
	fileNames := append([]string{stack.EntryPoint}, stack.AdditionalFiles...)
	manifestFilePaths := make([]string, len(fileNames))

	namespaces := map[string]bool{}
	if stack.Namespace != "" {
		namespaces[stack.Namespace] = true
	}

	tmpDir, err := ioutil.TempDir("", "kub_deployment")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create temp kub deployment directory")
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	for _, fileName := range fileNames {
		manifestFilePath := filesystem.JoinPaths(tmpDir, fileName)
		manifestContent, err := ioutil.ReadFile(filesystem.JoinPaths(stack.ProjectPath, fileName))
		if err != nil {
			return nil, cleanup, errors.Wrap(err, "failed to read manifest file")
		}

		if stack.IsComposeFormat {
			manifestContent, err = kubeDeployer.ConvertCompose(manifestContent)
			if err != nil {
				return nil, cleanup, errors.Wrap(err, "failed to convert docker compose file to a kube manifest")
			}
		}

		manifestContent, err = k.AddAppLabels(manifestContent, appLabels.ToMap())
		if err != nil {
			return nil, cleanup, errors.Wrap(err, "failed to add application labels")
		}

		// get resource namespace, fallback to provided namespace if not explicit on resource
		resourceNamespace, err := kubernetes.GetNamespace(manifestContent)
		if err != nil {
			return nil, cleanup, errors.Wrap(err, "failed to get resource namespace")
		}

		if resourceNamespace != "" {
			namespaces[resourceNamespace] = true
		}

		err = filesystem.WriteToFile(manifestFilePath, []byte(manifestContent))
		if err != nil {
			return nil, cleanup, errors.Wrap(err, "failed to create temp manifest file")
		}

		manifestFilePaths = append(manifestFilePaths, manifestFilePath)
	}

	namespacesList := make([]string, 0, len(namespaces))
	for namespace := range namespaces {
		namespacesList = append(namespacesList, namespace)
	}

	return &DeploymentFiles{
		FilePaths:  manifestFilePaths,
		Namespaces: namespacesList,
	}, cleanup, nil

}
