package cli

import (
	"context"
	"strings"

	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/concurrent"
	"github.com/portainer/portainer-ee/api/internal/errorlist"
	"github.com/rs/zerolog/log"

	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kcl *KubeClient) GetClusterRoles() ([]models.K8sClusterRole, error) {

	listClusterRoles := func(ctx context.Context) (interface{}, error) {
		return kcl.cli.RbacV1().ClusterRoles().List(ctx, meta.ListOptions{})
	}

	listClusterRoleBindings := func(ctx context.Context) (interface{}, error) {
		return kcl.cli.RbacV1().ClusterRoleBindings().List(ctx, meta.ListOptions{})
	}

	listNamespaces := func(ctx context.Context) (interface{}, error) {
		return kcl.cli.CoreV1().Namespaces().List(ctx, meta.ListOptions{})
	}

	results, err := concurrent.Run(context.TODO(), listClusterRoles, listClusterRoleBindings, listNamespaces)
	if err != nil {
		return nil, err
	}

	var clusterRoleList *rbac.ClusterRoleList
	var clusterRoleBindingList *rbac.ClusterRoleBindingList
	var namespaceList *core.NamespaceList
	for _, r := range results {
		switch v := r.Result.(type) {
		case *rbac.ClusterRoleList:
			clusterRoleList = v
		case *rbac.ClusterRoleBindingList:
			clusterRoleBindingList = v
		case *core.NamespaceList:
			namespaceList = v
		}
	}

	var roles []models.K8sClusterRole
	for _, r := range clusterRoleList.Items {
		role := models.K8sClusterRole{
			Name:            r.Name,
			Namespace:       r.Namespace,
			UID:             r.UID,
			ResourceVersion: r.ResourceVersion,
			Annotations:     r.Annotations,
			CreationDate:    r.CreationTimestamp.Time,
			Rules:           r.Rules,
			IsUnused:        true,
			IsSystem:        isSystemClusterRole(&r),
		}
		roles = append(roles, role)
	}

	// TODO
	// Can we create a field or label selector to exclude system resources when asked? This would be a lot more efficient.
	// Sample code below on how we might be able to do it, either check for Namespace or use the fieldselector or labelselector

	// // create a label selector with the "not" operator
	// selector, err := labels.Parse("kubernetes.io/bootstrapping!=rbac-defaults")
	// if err != nil {
	// 	panic(err.Error())
	// }

	// mark clusterRoles that are not used by any cluster role binding
	for _, crb := range clusterRoleBindingList.Items {
		for i, role := range roles {
			if role.Name == crb.RoleRef.Name {
				roles[i].IsUnused = false
			}
		}
	}

	// Cluster roles can also be referenced by rolebindings also
	for _, namespace := range namespaceList.Items {
		roleBindingList, err := kcl.cli.RbacV1().RoleBindings(namespace.Name).List(context.Background(), meta.ListOptions{})
		if err != nil {
			return roles, nil
		}

		// then mark the roles that are used by a role binding
		for _, roleBinding := range roleBindingList.Items {
			for i, role := range roles {
				if role.Name == roleBinding.RoleRef.Name {
					roles[i].IsUnused = false
				}
			}
		}
	}

	return roles, err
}

func isSystemClusterRole(role *rbac.ClusterRole) bool {
	if role.Namespace == "kube-system" || role.Namespace == "kube-public" ||
		role.Namespace == "kube-node-lease" || role.Namespace == "portainer" {
		return true
	}

	if strings.HasPrefix(role.Name, "system:") {
		return true
	}

	if role.Labels != nil {
		if role.Labels["kubernetes.io/bootstrapping"] == "rbac-defaults" {
			return true
		}
	}

	roles := getPortainerDefaultK8sRoleNames()
	for i := range roles {
		if role.Name == roles[i] {
			return true
		}
	}

	return false
}

func (kcl *KubeClient) DeleteClusterRoles(req models.K8sClusterRoleDeleteRequests) error {
	var errors []error
	for _, name := range req {
		client := kcl.cli.RbacV1().ClusterRoles()

		clusterRole, err := client.Get(context.Background(), name, meta.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			// this is a more serious error to do with the client so we return right away
			return err
		}

		if isSystemClusterRole(clusterRole) {
			log.Warn().Msgf("Ignoring delete of 'system' cluster role %q. Not allowed", name)
		}

		err = client.Delete(context.Background(), name, meta.DeleteOptions{})
		if err != nil {
			log.Err(err).Msgf("unable to delete the cluster role named: %q", name)
			errors = append(errors, err)
		}
	}

	return errorlist.Combine(errors)
}
