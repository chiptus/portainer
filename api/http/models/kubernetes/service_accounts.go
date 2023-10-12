package models

import (
	"errors"
	"net/http"
	"time"
)

type (
	K8sServiceAccount struct {
		Name         string    `json:"name"`
		UID          string    `json:"uid"`
		Namespace    string    `json:"namespace"`
		CreationDate time.Time `json:"creationDate"`

		IsSystem bool `json:"isSystem"`
		IsUnused bool `json:"isUnused"`
	}

	// K8sServiceAcountDeleteRequests is a mapping of namespace names to a slice of service account names.
	K8sServiceAccountDeleteRequests map[string][]string
)

func (r K8sServiceAccountDeleteRequests) Validate(request *http.Request) error {
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
