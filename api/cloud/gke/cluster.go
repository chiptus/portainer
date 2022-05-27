package gke

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

type Config struct {
	APIVersion     string `yaml:"apiVersion"`
	Clusters       []ConfigCluster
	Contexts       []ConfigContext
	CurrentContext string   `yaml:"current-context"`
	Kind           string   `yaml:"kind"`
	Preferences    struct{} `yaml:"preferences"`
	Users          []ConfigUsers
}

type ConfigCluster struct {
	Cluster struct {
		CertificateAuthorityData string `yaml:"certificate-authority-data"`
		Server                   string `yaml:"server"`
	} `yaml:"cluster"`
	Name string `yaml:"name"`
}

type ConfigContext struct {
	Context struct {
		Cluster string `yaml:"cluster"`
		User    string `yaml:"user"`
	} `yaml:"context"`
	Name string `yaml:"name"`
}

type ConfigUsers struct {
	Name string `yaml:"name"`
	User ConfigUser
}

type ConfigUser struct {
	AuthProvider struct {
		Name string `yaml:"name"`
	} `yaml:"auth-provider"`
}

// BuildConfig builds a KubeConfig for a specific cluster by requesting the
// relevant information from GKE.
func (k Key) BuildConfig(ctx context.Context, clusterID string) ([]byte, error) {
	tmp := strings.Split(clusterID, ":")
	if len(tmp) != 2 {
		return nil, fmt.Errorf("clusterID is not valid: it must be in the form zone:name, but it is %v", clusterID)
	}
	zone := tmp[0]
	clusterName := tmp[1]

	// Basic config structure.
	config := Config{
		APIVersion: "v1",
		Kind:       "Config",
	}

	// Use our single cluster to build the "list" in the KubeConfig
	cluster, err := k.FindCluster(ctx, zone, clusterName)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("gke_%s_%s_%s", k.ProjectID, cluster.Zone, cluster.Name)
	cert := cluster.MasterAuth.ClusterCaCertificate
	if err != nil {
		return nil, fmt.Errorf(
			"invalid certificate cluster=%s cert=%s: %w",
			name,
			cluster.MasterAuth.ClusterCaCertificate,
			err,
		)
	}

	var clusters []ConfigCluster
	clusters = append(clusters, ConfigCluster{
		Name: name,
		Cluster: struct {
			CertificateAuthorityData string "yaml:\"certificate-authority-data\""
			Server                   string "yaml:\"server\""
		}{
			CertificateAuthorityData: string(cert),
			Server:                   "https://" + cluster.Endpoint,
		},
	})
	config.Clusters = clusters

	// Context is the same since we're only building a KubeConfig with a single
	// cluster.
	var contexts []ConfigContext
	contexts = append(contexts, ConfigContext{
		Name: name,
		Context: struct {
			Cluster string "yaml:\"cluster\""
			User    string "yaml:\"user\""
		}{
			Cluster: name,
			User:    name,
		},
	})
	config.Contexts = contexts
	config.CurrentContext = name

	// Finally, the users section. It's simple, but has a weird structure.
	var users []ConfigUsers
	users = append(users, ConfigUsers{
		Name: name,
		User: ConfigUser{
			AuthProvider: struct {
				Name string "yaml:\"name\""
			}{Name: "gcp"},
		},
	})
	config.Users = users

	configYAML, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	return configYAML, nil
}

func (k Key) FindCluster(ctx context.Context, zone, name string) (*container.Cluster, error) {
	svc, err := container.NewService(
		ctx,
		option.WithCredentialsJSON(k.Bytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating container.NewService: %w", err)
	}

	// Ask Google for a list of all kube clusters in the given project in the
	// given zone.
	resp, err := svc.Projects.Zones.Clusters.List(k.ProjectID, zone).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("clusters list project=%s: %w", k.ProjectID, err)
	}
	for _, cluster := range resp.Clusters {
		// See if the cluster exists.
		if cluster.Name == name {
			if cluster.Status != "RUNNING" {
				return nil, fmt.Errorf("cluster is %s", cluster.Status)
			}
			if cluster.Endpoint == "" {
				return nil, fmt.Errorf("cluster does not have an IP address yet")
			}
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("cluster %s not available yet", name)
}
