package models

type (
	K8sApplication struct {
		UID       string
		Name      string
		Namespace string
		Kind      string
		Labels    map[string]string
	}
)
