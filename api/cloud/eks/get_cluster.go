package eks

import (
	"os"

	"github.com/portainer/portainer-ee/api/cloud/eks/eksctl"
	clouderrors "github.com/portainer/portainer-ee/api/cloud/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type (
	Cluster struct {
		Name       string
		Status     string
		Ready      bool
		KubeConfig string
	}

	Config struct {
		APIVersion     string `yaml:"apiVersion"`
		Clusters       []ConfigCluster
		Contexts       []ConfigContext
		CurrentContext string   `yaml:"current-context"`
		Kind           string   `yaml:"kind"`
		Preferences    struct{} `yaml:"preferences"`
		Users          []ConfigUsers
	}

	ConfigCluster struct {
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
			Server                   string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	}

	ConfigContext struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	}

	ConfigUsers struct {
		Name string     `yaml:"name"`
		User ConfigUser `yaml:"user"`
	}

	ConfigUser struct {
		Exec struct {
			APIVersion         string    `yaml:"apiVersion"`
			Command            string    `yaml:"command"`
			Args               []string  `yaml:"args"`
			Env                []ExecEnv `yaml:"env"`
			ProvideClusterInfo bool      `yaml:"provideClusterInfo"`
		} `yaml:"exec"`
	}

	ExecEnv struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	}
)

func (e *EksProvisioner) GetCluster(name string) (*Cluster, error) {
	log.Debugf("[cloud] [message: sending KaaS cluster details request] [provider: Amazon EKS] [clusterName: %s] [region: %s]", name, e.Region)

	cfg := eksctl.NewConfig(name, e.AccessKeyId, e.SecretAccessKey, e.Region, e.BinaryPath)

	kubeconfig, err := os.CreateTemp("", "")
	if err != nil {
		return nil, clouderrors.NewFatalError("could not create temp file for kubeconfig, err: %v", err)
	}
	defer os.Remove(kubeconfig.Name())

	args := []string{"utils", "write-kubeconfig",
		"--cluster", name,
		"--region", e.Region,
		"--kubeconfig", kubeconfig.Name(),
	}

	err = cfg.Run(args...)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(kubeconfig.Name())
	if err != nil {
		return nil, clouderrors.NewFatalError("read kubeconfig failed, err: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return nil, clouderrors.NewFatalError("unmarshal kubeconfig failed, err: %v", err)
	}

	config.Users[0].User.Exec.Env = append(config.Users[0].User.Exec.Env,
		[]ExecEnv{
			{"AWS_ACCESS_KEY_ID", e.AccessKeyId},
			{"AWS_SECRET_ACCESS_KEY", e.SecretAccessKey},
		}...)

	b, err = yaml.Marshal(config)
	if err != nil {
		return nil, clouderrors.NewFatalError("marshal kubeconfig failed, err: %v", err)
	}

	cluster := &Cluster{
		Name:       name,
		Ready:      true,
		Status:     "ACTIVE",
		KubeConfig: string(b),
	}

	return cluster, nil
}
