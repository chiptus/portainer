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
)

func (kcl *KubeClient) GetClusterRoleBindings() ([]models.K8sClusterRoleBinding, error) {
	clusterRoleBindingList, err := kcl.cli.RbacV1().ClusterRoleBindings().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	bindings := make([]models.K8sClusterRoleBinding, len(clusterRoleBindingList.Items))
	for i, crb := range clusterRoleBindingList.Items {
		bindings[i] = models.K8sClusterRoleBinding{
			Name:            crb.Name,
			UID:             crb.UID,
			ResourceVersion: crb.ResourceVersion,
			Annotations:     crb.Annotations,
			CreationDate:    crb.CreationTimestamp.Time,

			RoleRef:  crb.RoleRef,
			Subjects: crb.Subjects,

			IsSystem: isSystemClusterRoleBinding(&crb),
		}
	}

	return bindings, err
}

func isSystemClusterRoleBinding(binding *corev1.ClusterRoleBinding) bool {
	if strings.HasPrefix(binding.Name, "system:") {
		return true
	}

	if binding.Labels != nil {
		if binding.Labels["kubernetes.io/bootstrapping"] == "rbac-defaults" {
			return true
		}
	}

	for _, sub := range binding.Subjects {
		if strings.HasPrefix(sub.Name, "system:") {
			return true
		}

		if sub.Namespace == "kube-system" ||
			sub.Namespace == "kube-public" ||
			sub.Namespace == "kube-node-lease" ||
			sub.Namespace == "portainer" {
			return true
		}
	}

	return false
}

// DeleteClusterRoleBindings processes a K8sClusterRoleBindingDeleteRequest
// by deleting each cluster role binding in its given namespace. If deleting a specific cluster role binding
// fails, the error is logged and we continue to delete the remaining cluster role bindings.
func (kcl *KubeClient) DeleteClusterRoleBindings(reqs models.K8sClusterRoleBindingDeleteRequests) error {
	var errors []error

	for _, name := range reqs {
		client := kcl.cli.RbacV1().ClusterRoleBindings()

		clusterRoleBinding, err := client.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			// this is a more serious error to do with the client so we return right away
			return err
		}

		if isSystemClusterRoleBinding(clusterRoleBinding) {
			log.Warn().Msgf("Ignoring delete of 'system' cluster role binding %q. Not allowed", name)
		}

		err = client.Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil {
			log.Err(err).Msgf("unable to delete the cluster role binding named: %q", name)
			errors = append(errors, err)
		}
	}

	return errorlist.Combine(errors)
}
