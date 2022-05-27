package gke

type ProvisionRequest struct {
	APIKey            string
	Zone              string
	ClusterName       string
	Subnet            string
	NodeSize          string
	CPU               int
	RAM               float64
	HDD               int
	NodeCount         int
	KubernetesVersion string
}
