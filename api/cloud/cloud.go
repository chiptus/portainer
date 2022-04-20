package cloud

type KaasCluster struct {
	Id         string `json:"Id"`
	Name       string `json:"Name"`
	Ready      bool   `json:"Ready"`
	KubeConfig string `json:"KubeConfig"`
}
