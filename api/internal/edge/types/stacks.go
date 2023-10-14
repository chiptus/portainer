package types

import portainer "github.com/portainer/portainer/api"

type StoreManifestFunc func(stackFolder string, relatedEndpointIds []portainer.EndpointID) (configPath, manifestPath, projectPath string, err error)
