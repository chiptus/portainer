package endpoints

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateKubeConfigEnvironment(t *testing.T) {
	type test struct {
		input []byte
		want  string
		err   error
	}

	tests := []test{
		{
			input: []byte(`apiVersion: v1`),
			want:  "YXBpVmVyc2lvbjogdjE=",
			err:   nil,
		},
		{
			input: []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: base64string
    extensions:
    - extension:
        last-update: Tue, 14 Jun 2022 14:01:31 PDT
        provider: minikube.sigs.k8s.io
        version: v1.25.2
      name: cluster_info
    server: https://172.20.28.196:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    extensions:
    - extension:
        last-update: Tue, 14 Jun 2022 14:01:31 PDT
        provider: minikube.sigs.k8s.io
        version: v1.25.2
      name: context_info
    namespace: default
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate-data: base64string
    client-key-data: base64string
`),
			want: "CmFwaVZlcnNpb246IHYxCmNsdXN0ZXJzOgotIGNsdXN0ZXI6CiAgICBjZXJ0aWZpY2F0ZS1hdXRob3JpdHktZGF0YTogYmFzZTY0c3RyaW5nCiAgICBleHRlbnNpb25zOgogICAgLSBleHRlbnNpb246CiAgICAgICAgbGFzdC11cGRhdGU6IFR1ZSwgMTQgSnVuIDIwMjIgMTQ6MDE6MzEgUERUCiAgICAgICAgcHJvdmlkZXI6IG1pbmlrdWJlLnNpZ3MuazhzLmlvCiAgICAgICAgdmVyc2lvbjogdjEuMjUuMgogICAgICBuYW1lOiBjbHVzdGVyX2luZm8KICAgIHNlcnZlcjogaHR0cHM6Ly8xNzIuMjAuMjguMTk2Ojg0NDMKICBuYW1lOiBtaW5pa3ViZQpjb250ZXh0czoKLSBjb250ZXh0OgogICAgY2x1c3RlcjogbWluaWt1YmUKICAgIGV4dGVuc2lvbnM6CiAgICAtIGV4dGVuc2lvbjoKICAgICAgICBsYXN0LXVwZGF0ZTogVHVlLCAxNCBKdW4gMjAyMiAxNDowMTozMSBQRFQKICAgICAgICBwcm92aWRlcjogbWluaWt1YmUuc2lncy5rOHMuaW8KICAgICAgICB2ZXJzaW9uOiB2MS4yNS4yCiAgICAgIG5hbWU6IGNvbnRleHRfaW5mbwogICAgbmFtZXNwYWNlOiBkZWZhdWx0CiAgICB1c2VyOiBtaW5pa3ViZQogIG5hbWU6IG1pbmlrdWJlCmN1cnJlbnQtY29udGV4dDogbWluaWt1YmUKa2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQp1c2VyczoKLSBuYW1lOiBtaW5pa3ViZQogIHVzZXI6CiAgICBjbGllbnQtY2VydGlmaWNhdGUtZGF0YTogYmFzZTY0c3RyaW5nCiAgICBjbGllbnQta2V5LWRhdGE6IGJhc2U2NHN0cmluZwo=",
			err:  nil,
		},
		{
			input: []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: C:\Users\kota\.minikube\ca.crt
  name: cert-auth-fail`),
			want: "",
			err: &SelfContainedConfigFileError{
				File: `C:\Users\kota\.minikube\ca.crt`,
			},
		},
		{
			input: []byte(`
apiVersion: v1
users:
- name: minikube
  user:
    client-certificate: C:\Users\kota\.minikube\profiles\minikube\client.crt`),
			want: "",
			err: &SelfContainedConfigFileError{
				File: `C:\Users\kota\.minikube\profiles\minikube\client.crt`,
			},
		},
		{
			input: []byte(`
apiVersion: v1
users:
- name: minikube
  user:
    client-key: C:\Users\kota\.minikube\profiles\minikube\client.key`),
			want: "",
			err: &SelfContainedConfigFileError{
				File: `C:\Users\kota\.minikube\profiles\minikube\client.key`,
			},
		},
		{
			input: []byte(`
users:
- name: gke_sublime-seat-351021_us-central1-c_cluster-1
  user:
    auth-provider:
      config:
        cmd-args: config config-helper --format=json
      name: gcp
`),
			want: "",
			err:  &SelfContainedConfigExecError{},
		},
		{
			input: []byte(`
users:
- name: gke_sublime-seat-351021_us-central1-c_cluster-1
  user:
    auth-provider:
      config:
        cmd-path: /Users/work/Downloads/google-cloud-sdk/bin/gcloud
      name: gcp
`),
			want: "",
			err:  &SelfContainedConfigExecError{},
		},
		{
			input: []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: C:\Users\kota\.minikube\ca.crt
    server: https://172.20.28.196:8443
  name: minikube
- cluster:
    certificate-authority-data: base64string
    server: https://35.202.10.252
  name: gke
contexts:
- context:
    cluster: minikube
    extensions:
    - extension:
        last-update: Tue, 14 Jun 2022 14:01:31 PDT
        provider: minikube.sigs.k8s.io
        version: v1.25.2
      name: context_info
    namespace: default
    user: minikube
  name: minikube
- context:
    cluster: gke
    user: gke
  name: gke
current-context: gke
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate: C:\Users\kota\.minikube\profiles\minikube\client.crt
    client-key: C:\Users\kota\.minikube\profiles\minikube\client.key
- name: gke
  user:
    auth-provider:
      name: gcp
`),
			want: "CmFwaVZlcnNpb246IHYxCmNsdXN0ZXJzOgotIGNsdXN0ZXI6CiAgICBjZXJ0aWZpY2F0ZS1hdXRob3JpdHk6IEM6XFVzZXJzXGtvdGFcLm1pbmlrdWJlXGNhLmNydAogICAgc2VydmVyOiBodHRwczovLzE3Mi4yMC4yOC4xOTY6ODQ0MwogIG5hbWU6IG1pbmlrdWJlCi0gY2x1c3RlcjoKICAgIGNlcnRpZmljYXRlLWF1dGhvcml0eS1kYXRhOiBiYXNlNjRzdHJpbmcKICAgIHNlcnZlcjogaHR0cHM6Ly8zNS4yMDIuMTAuMjUyCiAgbmFtZTogZ2tlCmNvbnRleHRzOgotIGNvbnRleHQ6CiAgICBjbHVzdGVyOiBtaW5pa3ViZQogICAgZXh0ZW5zaW9uczoKICAgIC0gZXh0ZW5zaW9uOgogICAgICAgIGxhc3QtdXBkYXRlOiBUdWUsIDE0IEp1biAyMDIyIDE0OjAxOjMxIFBEVAogICAgICAgIHByb3ZpZGVyOiBtaW5pa3ViZS5zaWdzLms4cy5pbwogICAgICAgIHZlcnNpb246IHYxLjI1LjIKICAgICAgbmFtZTogY29udGV4dF9pbmZvCiAgICBuYW1lc3BhY2U6IGRlZmF1bHQKICAgIHVzZXI6IG1pbmlrdWJlCiAgbmFtZTogbWluaWt1YmUKLSBjb250ZXh0OgogICAgY2x1c3RlcjogZ2tlCiAgICB1c2VyOiBna2UKICBuYW1lOiBna2UKY3VycmVudC1jb250ZXh0OiBna2UKa2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQp1c2VyczoKLSBuYW1lOiBtaW5pa3ViZQogIHVzZXI6CiAgICBjbGllbnQtY2VydGlmaWNhdGU6IEM6XFVzZXJzXGtvdGFcLm1pbmlrdWJlXHByb2ZpbGVzXG1pbmlrdWJlXGNsaWVudC5jcnQKICAgIGNsaWVudC1rZXk6IEM6XFVzZXJzXGtvdGFcLm1pbmlrdWJlXHByb2ZpbGVzXG1pbmlrdWJlXGNsaWVudC5rZXkKLSBuYW1lOiBna2UKICB1c2VyOgogICAgYXV0aC1wcm92aWRlcjoKICAgICAgbmFtZTogZ2NwCg==",
			err:  nil,
		},
		{
			input: []byte(``),
			want:  "",
			err:   errors.New("Missing or invalid kubeconfig"),
		},
		{
			input: []byte("\t"),
			want:  "",
			err:   errors.New("KubeConfig could not be parsed as yaml"),
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			config := base64.StdEncoding.EncodeToString(tc.input)
			r := http.Request{
				Form: url.Values{
					"KubeConfig": []string{config},
				},
			}

			got, err := validateKubeConfigEnvironment(&r)
			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
