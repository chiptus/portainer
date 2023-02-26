package types

import portaineree "github.com/portainer/portainer-ee/api"

type StoreManifestFunc func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (configPath, manifestPath, projectPath string, err error)
