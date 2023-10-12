package models

import (
	"errors"
	"net/http"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

type (
	K8sClusterRole struct {
		Name            string            `json:"name"`
		UID             types.UID         `json:"uid"`
		Namespace       string            `json:"namespace"`
		ResourceVersion string            `json:"resourceVersion"`
		CreationDate    time.Time         `json:"creationDate"`
		Annotations     map[string]string `json:"annotations"`

		Rules []rbacv1.PolicyRule `json:"rules"`

		IsUnused bool `json:"isUnused"`
		IsSystem bool `json:"isSystem"`
	}

	K8sClusterRoleDeleteRequests []string
)

func (r K8sClusterRoleDeleteRequests) Validate(request *http.Request) error {
	if len(r) == 0 {
		return errors.New("missing deletion request list in payload")
	}

	return nil
}
