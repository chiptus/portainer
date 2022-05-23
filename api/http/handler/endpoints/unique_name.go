package endpoints

import portaineree "github.com/portainer/portainer-ee/api"

func (handler *Handler) isNameUnique(name string, endpointID portaineree.EndpointID) (bool, error) {
	endpoints, err := handler.dataStore.Endpoint().Endpoints()
	if err != nil {
		return false, err
	}

	for _, endpoint := range endpoints {
		if endpoint.Name == name && (endpointID == 0 || endpoint.ID != endpointID) {
			return false, nil
		}
	}

	return true, nil
}
