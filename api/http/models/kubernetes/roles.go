package models

import (
	"errors"
	"net/http"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

type (
	K8sRole struct {
		Name            string            `json:"name"`
		UID             types.UID         `json:"uid"`
		Namespace       string            `json:"namespace"`
		ResourceVersion string            `json:"resourceVersion"`
		CreationDate    time.Time         `json:"creationDate"`
		Annotations     map[string]string `json:"annotations"`

		Rules []rbacv1.PolicyRule `json:"rules"`

		// Unused is true if the role is not bound to any subject.
		IsUnused bool `json:"isUnused"`

		// isSystem is true if prefixed with "system:" or exists in the kube-system namespace
		// or is one of the portainer roles
		IsSystem bool `json:"isSystem"`
	}

	// K8sRoleDeleteRequests is a mapping of namespace names to a slice of roles.
	K8sRoleDeleteRequests map[string][]string
)

func (r K8sRoleDeleteRequests) Validate(request *http.Request) error {
	if len(r) == 0 {
		return errors.New("missing deletion request list in payload")
	}
	for ns := range r {
		if len(ns) == 0 {
			return errors.New("deletion given with empty namespace")
		}
	}
	return nil
}
