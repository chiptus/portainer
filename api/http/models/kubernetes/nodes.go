package models

import "net/http"

type K8sNodes struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

func (r *K8sNodes) Validate(request *http.Request) error {
	return nil
}
