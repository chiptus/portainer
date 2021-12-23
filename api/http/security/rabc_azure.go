package security

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"net/http"
	"path"
	"strings"
)

func getAzureOperationAuthorization(url, method string) portaineree.Authorization {
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
func azureSubscriptionsOperationAuthorization(url, method string) portaineree.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureSubscriptionsList
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*
func azureSubscriptionOperationAuthorization(url, method string) portaineree.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureSubscriptionGet
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/resourcegroups
func azureResourceGroupsOperationAuthorization(url, method string) portaineree.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureResourceGroupsList
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/resourcegroups/*
func azureResourceGroupOperationAuthorization(url, method string) portaineree.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureResourceGroupGet
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/providers/*
func azureProviderOperationAuthorization(url, method string) portaineree.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureProviderGet
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/providers/Microsoft.ContainerInstance/containerGroups
func azureContainerGroupsOperationAuthorization(url, method string) portaineree.Authorization {
	switch method {
	case http.MethodGet:
		return portaineree.OperationAzureContainerGroupsList
	default:
		return portaineree.OperationAzureUndefined
	}
}

// /subscriptions/*/resourceGroups/*/providers/Microsoft.ContainerInstance/containerGroups/*
func azureContainerGroupOperationAuthorization(url, method string) portaineree.Authorization {
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
