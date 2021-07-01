package registries

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
)

func Test_getRegistryManagementUrl(t *testing.T) {
	type args struct {
		registry *portainer.Registry
	}

	customRegistryUrl := "https://example.com"
	baseUrl := customRegistryUrl
	registryWithFeedUrl := customRegistryUrl + "/dev"
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ProGet registry without baseUrl",
			args: args{
				registry: &portainer.Registry{
					Type: portainer.ProGetRegistry,
					URL:  registryWithFeedUrl,
				},
			},
			want: registryWithFeedUrl,
		},
		{
			name: "ProGet registry with baseUrl",
			args: args{
				registry: &portainer.Registry{
					Type:    portainer.ProGetRegistry,
					URL:     registryWithFeedUrl,
					BaseURL: baseUrl,
				},
			},
			want: baseUrl,
		},
		{
			name: "Custom registry - no baseUrl",
			args: args{
				registry: &portainer.Registry{
					Type: portainer.CustomRegistry,
					URL:  customRegistryUrl,
				},
			},
			want: customRegistryUrl,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRegistryManagementUrl(tt.args.registry); got != tt.want {
				t.Errorf("getRegistryManagementUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
