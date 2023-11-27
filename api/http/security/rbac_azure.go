package security

import (
	"net/http"
	"path"
	"strings"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

func getAzureOperationAuthorization(url, method string) portainer.Authorization {
	url = strings.Split(url, "?")[0]
	if matched, _ := path.Match("/subscriptions", url); matched {
		return azureSubscriptionsOperationAuthorization(url, method)
	} else if matched, _ := path.Match("/subscriptions/*", url); matched {
		return azureSubscriptionOperationAuthorization(url, method)
	} else if matched, _ := path.Match("/subscriptions/*/providers/*", url); matched {
		return azureProviderOperationAuthorization(url, method)
	} else if matched, _ := path.Match("/subscriptions/*/resourcegroups", url); matched {
		return azureResourceGroupsOperationAuthorization(url, method)
	} else if matched, _ := path.Match("/subscriptions/*/resourcegroups/*", url); matched {
		return azureResourceGroupOperationAuthorization(url, method)
	} else if matched, _ := path.Match("/subscriptions/*/providers/*/containerGroups", url); matched {
		return azureContainerGroupsOperationAuthorization(url, method)
	} else if matched, _ := path.Match("/subscriptions/*/resourceGroups/*/providers/*/containerGroups/*", url); matched {
		return azureContainerGroupOperationAuthorization(url, method)
	}

	return portaineree.OperationAzureUndefined
}

// /subscriptions
func azureSubscriptionsOperationAuthorization(url, method string) portainer.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureSubscriptionsList
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*
func azureSubscriptionOperationAuthorization(url, method string) portainer.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureSubscriptionGet
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/resourcegroups
func azureResourceGroupsOperationAuthorization(url, method string) portainer.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureResourceGroupsList
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/resourcegroups/*
func azureResourceGroupOperationAuthorization(url, method string) portainer.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureResourceGroupGet
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/providers/*
func azureProviderOperationAuthorization(url, method string) portainer.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureProviderGet
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/providers/Microsoft.ContainerInstance/containerGroups
func azureContainerGroupsOperationAuthorization(url, method string) portainer.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureContainerGroupsList
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/resourceGroups/*/providers/Microsoft.ContainerInstance/containerGroups/*
func azureContainerGroupOperationAuthorization(url, method string) portainer.Authorization {
	switch method {
	case http.MethodPut:
		return portaineree.OperationAzureContainerGroupCreate
	case http.MethodGet:
		return portaineree.OperationAzureContainerGroupGet
	case http.MethodDelete:
		return portaineree.OperationAzureContainerGroupDelete
	default:
		return portaineree.OperationAzureUndefined
	}
}