package cli

import (
	"context"
	"strings"

	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	"github.com/portainer/portainer-ee/api/internal/errorlist"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kcl *KubeClient) GetRoleBindings(namespace string) ([]models.K8sRoleBinding, error) {
	roleBindingList, err := kcl.cli.RbacV1().RoleBindings(namespace).List(context.Background(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	bindings := make([]models.K8sRoleBinding, len(roleBindingList.Items))
	for i, rb := range roleBindingList.Items {
		bindings[i] = models.K8sRoleBinding{
			Name:            rb.Name,
			Namespace:       rb.Namespace,
			UID:             rb.UID,
			ResourceVersion: rb.ResourceVersion,
			Annotations:     rb.Annotations,
			CreationDate:    rb.CreationTimestamp.Time,

			RoleRef:  rb.RoleRef,
			Subjects: rb.Subjects,

			IsSystem: kcl.isSystemRoleBinding(&rb),
		}
	}

	return bindings, err
}

func (kcl *KubeClient) isSystemRoleBinding(rb *corev1.RoleBinding) bool {
	if strings.HasPrefix(rb.Name, "system:") {
		return true
	}

	if rb.Labels != nil {
		if rb.Labels["kubernetes.io/bootstrapping"] == "rbac-defaults" {
			return true
		}
	}

	if rb.RoleRef.Name != "" {
		role, err := kcl.getRole(rb.Namespace, rb.RoleRef.Name)
		if err != nil {
			return false
		}

		// Linked to a role that is marked a system role
		if kcl.isSystemRole(role) {
			return true
		}
	}

	return false
}

func (kcl *KubeClient) getRole(namespace, name string) (*corev1.Role, error) {
	client := kcl.cli.RbacV1().Roles(namespace)
	return client.Get(context.Background(), name, metav1.GetOptions{})
}

// DeleteRoleBindings processes a K8sServiceDeleteRequest by deleting each service
// in its given namespace.
func (kcl *KubeClient) DeleteRoleBindings(reqs models.K8sRoleBindingDeleteRequests) error {
	var errors []error
	for namespace := range reqs {
		for _, name := range reqs[namespace] {
			client := kcl.cli.RbacV1().RoleBindings(namespace)

			roleBinding, err := client.Get(context.Background(), name, v1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				// this is a more serious error to do with the client so we return right away
				return err
			}

			if kcl.isSystemRoleBinding(roleBinding) {
				log.Error().Msgf("Ignoring delete of 'system' role binding %q. Not allowed", name)
			}

			err = client.Delete(context.Background(), name, v1.DeleteOptions{})
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	return errorlist.Combine(errors)
}
