package cli

import (
	"context"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceAccount returns the portainer ServiceAccount associated to the specified user.
func (kcl *KubeClient) GetServiceAccount(tokenData *portaineree.TokenData) (*v1.ServiceAccount, error) {
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
	endpointRoleID portaineree.RoleID,
	namespaces map[string]portaineree.K8sNamespaceInfo,
	namespaceRoles map[string]portaineree.Role,
) error {
	serviceAccountName := UserServiceAccountName(int(user.ID), kcl.instanceID)

	err := kcl.upsertPortainerK8sClusterRoles()
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
	endpointRoleID portaineree.RoleID,
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
