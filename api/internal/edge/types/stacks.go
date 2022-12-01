package types

import portaineree "github.com/portainer/portainer-ee/api"

type StoreManifestFunc func(stackFolder string, relatedEndpointIds []portaineree.EndpointID) (composePath, manifestPath, projectPath string, err error)
