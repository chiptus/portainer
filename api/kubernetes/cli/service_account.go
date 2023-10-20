package cli

import (
	"context"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/concurrent"
	"github.com/portainer/portainer-ee/api/internal/errorlist"
	portainer "github.com/portainer/portainer/api"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
)

// GetServiceAccount returns the portainer ServiceAccount associated to the specified user.
func (kcl *KubeClient) GetServiceAccount(tokenData *portainer.TokenData) (*v1.ServiceAccount, error) {
	var portainerServiceAccountName string
	if tokenData.Role == portaineree.AdministratorRole {
		portainerServiceAccountName = portainerClusterAdminServiceAccountName
	} else {
		portainerServiceAccountName = UserServiceAccountName(int(tokenData.ID), kcl.instanceID)
	}

	// verify name exists as service account resource within portainer namespace
	serviceAccount, err := kcl.cli.CoreV1().ServiceAccounts(portainerNamespace).Get(context.TODO(), portainerServiceAccountName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return serviceAccount, nil
}

// GetServiceAccounts returns a list of ServiceAccounts in the given namespace
func (kcl *KubeClient) GetServiceAccounts(namespace string) ([]models.K8sServiceAccount, error) {
	serviceAccountList, err := kcl.cli.CoreV1().ServiceAccounts(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	serviceAccounts := make([]models.K8sServiceAccount, len(serviceAccountList.Items))
	for i := 0; i < len(serviceAccountList.Items); i++ {
		item := &serviceAccountList.Items[i]
		sa := models.K8sServiceAccount{
			Name:         item.Name,
			Namespace:    item.Namespace,
			CreationDate: item.CreationTimestamp.Time,
			UID:          string(item.UID),
		}

		serviceAccounts[i] = sa
	}

	kcl.lookupSystemResources(serviceAccountList, serviceAccounts)
	kcl.lookupUnusedResources(serviceAccountList, serviceAccounts)

	return serviceAccounts, nil
}

func (kcl *KubeClient) lookupSystemResources(serviceAccountList *v1.ServiceAccountList, serviceAccounts []models.K8sServiceAccount) {
	isSystemTask := func(sa *v1.ServiceAccount) concurrent.Func {
		return func(ctx context.Context) (interface{}, error) {
			result := kcl.isSystemServiceAccount(sa)
			return result, nil
		}
	}

	// Create a slice of tasks by iterating over the ServiceAccount pointers
	var tasks []concurrent.Func
	for i := 0; i < len(serviceAccountList.Items); i++ {
		taskFunc := isSystemTask(&serviceAccountList.Items[i])
		tasks = append(tasks, taskFunc)
	}

	// we can ignore errors here because the tasks here don't return errors
	results, _ := concurrent.Run(context.Background(), maxConcurrency, tasks...)

	for i, result := range results {
		// Update the ServiceAccount struct with the result
		serviceAccounts[i].IsSystem = result.Result.(bool)
	}
}

func (kcl *KubeClient) lookupUnusedResources(serviceAccountList *v1.ServiceAccountList, serviceAccounts []models.K8sServiceAccount) {
	isUnusedTask := func(sa *v1.ServiceAccount) concurrent.Func {
		return func(ctx context.Context) (interface{}, error) {
			result := kcl.isServiceAccountUnused(sa)
			return result, nil
		}
	}

	// Create a slice of tasks by iterating over the ServiceAccount pointers
	var tasks []concurrent.Func
	for _, sa := range serviceAccountList.Items {
		taskFunc := isUnusedTask(&sa)
		tasks = append(tasks, taskFunc)
	}

	// Run the tasks concurrently
	results, _ := concurrent.Run(context.Background(), maxConcurrency, tasks...)

	for i, result := range results {
		// Update the ServiceAccount struct with the result
		serviceAccounts[i].IsUnused = result.Result.(bool)
	}
}

func (kcl *KubeClient) isSystemServiceAccount(sa *v1.ServiceAccount) bool {
	return kcl.isSystemNamespace(sa.Namespace)
}

func (kcl *KubeClient) isServiceAccountUnused(sa *v1.ServiceAccount) bool {
	selectors := map[string]string{
		"spec.serviceAccountName": sa.Name,
	}

	selectorLabels := labels.SelectorFromSet(selectors).String()

	// Check if service account is used by any pods
	podList, err := kcl.cli.CoreV1().Pods(sa.Namespace).List(context.TODO(),
		metav1.ListOptions{
			FieldSelector: selectorLabels,
			Limit:         1,
		},
	)

	if err != nil {
		return true
	}

	return len(podList.Items) == 0
}

// DeleteServices processes a K8sServiceDeleteRequest by deleting each service
// in its given namespace.
func (kcl *KubeClient) DeleteServiceAccounts(reqs models.K8sServiceAccountDeleteRequests) error {
	var errors []error
	for namespace := range reqs {
		for _, serviceName := range reqs[namespace] {
			client := kcl.cli.CoreV1().ServiceAccounts(namespace)

			sa, err := client.Get(context.Background(), serviceName, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return err
			}

			if kcl.isSystemServiceAccount(sa) {
				return fmt.Errorf("cannot delete system service account %q", namespace+"/"+serviceName)
			}

			err = client.Delete(context.Background(), serviceName, metav1.DeleteOptions{})
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errorlist.Combine(errors)
}

// GetServiceAccountBearerToken returns the ServiceAccountToken associated to the specified user.
func (kcl *KubeClient) GetServiceAccountBearerToken(userID int) (string, error) {
	serviceAccountName := UserServiceAccountName(userID, kcl.instanceID)

	return kcl.getServiceAccountToken(serviceAccountName)
}

// SetupUserServiceAccount will make sure that all the required resources are created inside the Kubernetes
// cluster before creating a ServiceAccount and a ServiceAccountToken for the specified Portainer user.
// It will also create required default RoleBinding and ClusterRoleBinding rules.
func (kcl *KubeClient) SetupUserServiceAccount(
	user portaineree.User,
	endpointRoleID portainer.RoleID,
	namespaces map[string]portaineree.K8sNamespaceInfo,
	namespaceRoles map[string]portaineree.Role,
	clusterConfig portaineree.KubernetesConfiguration,
) error {
	serviceAccountName := UserServiceAccountName(int(user.ID), kcl.instanceID)

	err := kcl.upsertPortainerK8sClusterRoles(clusterConfig)
	if err != nil {
		return err
	}

	err = kcl.createUserServiceAccount(portainerNamespace, serviceAccountName)
	if err != nil {
		return err
	}

	err = kcl.createServiceAccountToken(serviceAccountName)
	if err != nil {
		return err
	}

	err = kcl.ensureServiceAccountHasPortainerClusterRoles(
		serviceAccountName, user, endpointRoleID)
	if err != nil {
		return err
	}

	return kcl.ensureServiceAccountHasPortainerRoles(
		serviceAccountName, namespaces, namespaceRoles)
}

// RemoveUserBindings removes k8s bindings of a user in a namespace
func (kcl *KubeClient) RemoveUserNamespaceBindings(
	userID int,
	namespace string,
) error {
	serviceAccountName := UserServiceAccountName(userID, kcl.instanceID)

	err := kcl.removeRoleBinding(serviceAccountName, namespace)

	return err
}

// RemoveUserServiceAccount removes the service account and its
// role binding, cluster role binding.
func (kcl *KubeClient) RemoveUserServiceAccount(
	userID int,
) error {
	serviceAccountName := UserServiceAccountName(userID, kcl.instanceID)

	err := kcl.removeRoleBindings(serviceAccountName)
	if err != nil {
		return err
	}

	err = kcl.removeClusterRoleBindings(serviceAccountName)
	if err != nil {
		return err
	}

	err = kcl.removeUserServiceAccount(portainerNamespace, serviceAccountName)

	return err
}

func (kcl *KubeClient) createUserServiceAccount(namespace, serviceAccountName string) error {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountName,
		},
	}

	_, err := kcl.cli.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (kcl *KubeClient) removeUserServiceAccount(namespace, serviceAccountName string) error {
	err := kcl.cli.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), serviceAccountName, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}

// setup cluster role binding for a service account
func (kcl *KubeClient) ensureServiceAccountHasPortainerClusterRoles(
	serviceAccountName string,
	user portaineree.User,
	endpointRoleID portainer.RoleID,
) error {

	roleSet, ok := getPortainerK8sRoleMapping()[endpointRoleID]
	if !ok {
		return nil
	}

	kcl.removeClusterRoleBindings(serviceAccountName)

	for _, role := range roleSet.k8sClusterRoles {
		err := kcl.createClusterRoleBindings(serviceAccountName, string(role))
		if err != nil {
			return err
		}
	}

	return nil
}

// setup role binding for a service account
func (kcl *KubeClient) ensureServiceAccountHasPortainerRoles(
	serviceAccountName string,
	namespaces map[string]portaineree.K8sNamespaceInfo,
	namespaceRoles map[string]portaineree.Role,
) error {
	rolesMapping := getPortainerK8sRoleMapping()

	for namespace := range namespaces {
		// remove the namespace access from the service account
		err := kcl.removeRoleBinding(serviceAccountName, namespace)
		if err != nil {
			return err
		}

		// namespace roles should contain the default namespace access too
		nsRole, ok := namespaceRoles[namespace]
		if !ok {
			continue
		}

		debug := ""
		for ns, r := range namespaceRoles {
			debug = fmt.Sprintf("%s%s:%s;", debug, ns, r.Name)
		}

		// setup k8s role bindings for the namespace based on user's namespace role
		roleSet := rolesMapping[nsRole.ID]
		for _, role := range roleSet.k8sRoles {
			err = kcl.createRoleBinding(serviceAccountName, string(role), namespace, true)
			if err != nil && !k8serrors.IsAlreadyExists(err) {
				return err
			}
		}
	}

	return nil
}
