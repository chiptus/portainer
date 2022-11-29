package kubernetes

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AddAppLabels(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{
			name: "single deployment without labels",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
`,
			wantOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
`,
		},
		{
			name: "single deployment with existing labels",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    foo: bar
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
`,
			wantOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    foo: bar
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
`,
		},
		{
			name: "complex kompose output",
			input: `apiVersion: v1
items:
  - apiVersion: v1
    kind: Service
    metadata:
      creationTimestamp: null
      labels:
        io.kompose.service: web
      name: web
    spec:
      ports:
        - name: "5000"
          port: 5000
          targetPort: 5000
      selector:
        io.kompose.service: web
    status:
      loadBalancer: {}
  - apiVersion: apps/v1
    kind: Deployment
    metadata:
      creationTimestamp: null
      labels:
        io.kompose.service: redis
      name: redis
    spec:
      replicas: 1
      selector:
        matchLabels:
          io.kompose.service: redis
      strategy: {}
      template:
        metadata:
          creationTimestamp: null
          labels:
            io.kompose.service: redis
    status: {}
  - apiVersion: apps/v1
    kind: Deployment
    metadata:
      creationTimestamp: null
      name: web
    spec:
      replicas: 1
      selector:
        matchLabels:
          io.kompose.service: web
      strategy:
        type: Recreate
      template:
        metadata:
          creationTimestamp: null
          labels:
            io.kompose.service: web
    status: {}
kind: List
metadata: {}
`,
			wantOutput: `apiVersion: v1
items:
  - apiVersion: v1
    kind: Service
    metadata:
      creationTimestamp: null
      labels:
        io.kompose.service: web
        io.portainer.kubernetes.application.kind: git
        io.portainer.kubernetes.application.name: best-name
        io.portainer.kubernetes.application.owner: best-owner
        io.portainer.kubernetes.application.stack: best-name
        io.portainer.kubernetes.application.stackid: "123"
      name: web
    spec:
      ports:
        - name: "5000"
          port: 5000
          targetPort: 5000
      selector:
        io.kompose.service: web
    status:
      loadBalancer: {}
  - apiVersion: apps/v1
    kind: Deployment
    metadata:
      creationTimestamp: null
      labels:
        io.kompose.service: redis
        io.portainer.kubernetes.application.kind: git
        io.portainer.kubernetes.application.name: best-name
        io.portainer.kubernetes.application.owner: best-owner
        io.portainer.kubernetes.application.stack: best-name
        io.portainer.kubernetes.application.stackid: "123"
      name: redis
    spec:
      replicas: 1
      selector:
        matchLabels:
          io.kompose.service: redis
      strategy: {}
      template:
        metadata:
          creationTimestamp: null
          labels:
            io.kompose.service: redis
    status: {}
  - apiVersion: apps/v1
    kind: Deployment
    metadata:
      creationTimestamp: null
      labels:
        io.portainer.kubernetes.application.kind: git
        io.portainer.kubernetes.application.name: best-name
        io.portainer.kubernetes.application.owner: best-owner
        io.portainer.kubernetes.application.stack: best-name
        io.portainer.kubernetes.application.stackid: "123"
      name: web
    spec:
      replicas: 1
      selector:
        matchLabels:
          io.kompose.service: web
      strategy:
        type: Recreate
      template:
        metadata:
          creationTimestamp: null
          labels:
            io.kompose.service: web
    status: {}
kind: List
metadata: {}
`,
		},
		{
			name: "multiple items separated by ---",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
---
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    io.kompose.service: web
  name: web
spec:
  ports:
    - name: "5000"
      port: 5000
      targetPort: 5000
  selector:
    io.kompose.service: web
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    foo: bar
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
`,
			wantOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
---
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    io.kompose.service: web
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
  name: web
spec:
  ports:
    - name: "5000"
      port: 5000
      targetPort: 5000
  selector:
    io.kompose.service: web
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    foo: bar
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
  name: busybox
spec:
  replicas: 3
  selector:
    matchLabels:
      app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - image: busybox
          name: busybox
`,
		},
		{
			name:       "empty",
			input:      "",
			wantOutput: "",
		},
		{
			name: "no only deployments",
			input: `apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    io.kompose.service: web
  name: web
spec:
  ports:
    - name: "5000"
      port: 5000
      targetPort: 5000
  selector:
    io.kompose.service: web
`,
			wantOutput: `apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    io.kompose.service: web
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
  name: web
spec:
  ports:
    - name: "5000"
      port: 5000
      targetPort: 5000
  selector:
    io.kompose.service: web
`,
		},
	}

	labels := KubeAppLabels{
		StackID:   123,
		StackName: "best-name",
		Owner:     "best-owner",
		Kind:      "git",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddAppLabels([]byte(tt.input), labels.ToMap())
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOutput, string(result))
		})
	}
}

func Test_AddAppLabels_HelmApp(t *testing.T) {
	labels := GetHelmAppLabels("best-name", "best-owner")

	tests := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{
			name: "bitnami nginx configmap",
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-test-server-block
  labels:
    app.kubernetes.io/name: nginx
    helm.sh/chart: nginx-9.5.4
    app.kubernetes.io/instance: nginx-test
    app.kubernetes.io/managed-by: Helm
data:
  server-blocks-paths.conf: |-
    include  "/opt/bitnami/nginx/conf/server_blocks/ldap/*.conf";
    include  "/opt/bitnami/nginx/conf/server_blocks/common/*.conf";
`,
			wantOutput: `apiVersion: v1
data:
  server-blocks-paths.conf: |-
    include  "/opt/bitnami/nginx/conf/server_blocks/ldap/*.conf";
    include  "/opt/bitnami/nginx/conf/server_blocks/common/*.conf";
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/instance: nginx-test
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: nginx
    helm.sh/chart: nginx-9.5.4
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
  name: nginx-test-server-block
`,
		},
		{
			name: "bitnami nginx service",
			input: `apiVersion: v1
kind: Service
metadata:
  name: nginx-test
  labels:
    app.kubernetes.io/name: nginx
    helm.sh/chart: nginx-9.5.4
    app.kubernetes.io/instance: nginx-test
    app.kubernetes.io/managed-by: Helm
spec:
  type: LoadBalancer
  externalTrafficPolicy: "Cluster"
  ports:
    - name: http
      port: 80
      targetPort: http
  selector:
    app.kubernetes.io/name: nginx
    app.kubernetes.io/instance: nginx-test
`,
			wantOutput: `apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/instance: nginx-test
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: nginx
    helm.sh/chart: nginx-9.5.4
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
  name: nginx-test
spec:
  externalTrafficPolicy: Cluster
  ports:
    - name: http
      port: 80
      targetPort: http
  selector:
    app.kubernetes.io/instance: nginx-test
    app.kubernetes.io/name: nginx
  type: LoadBalancer
`,
		},
		{
			name: "bitnami nginx deployment",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-test
  labels:
    app.kubernetes.io/name: nginx
    helm.sh/chart: nginx-9.5.4
    app.kubernetes.io/instance: nginx-test
    app.kubernetes.io/managed-by: Helm
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: nginx
      app.kubernetes.io/instance: nginx-test
  template:
    metadata:
      labels:
        app.kubernetes.io/name: nginx
        helm.sh/chart: nginx-9.5.4
        app.kubernetes.io/instance: nginx-test
        app.kubernetes.io/managed-by: Helm
    spec:
      automountServiceAccountToken: false
      shareProcessNamespace: false
      serviceAccountName: default
      containers:
        - name: nginx
          image: docker.io/bitnami/nginx:1.21.3-debian-10-r0
          imagePullPolicy: "IfNotPresent"
`,
			wantOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/instance: nginx-test
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: nginx
    helm.sh/chart: nginx-9.5.4
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
  name: nginx-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: nginx-test
      app.kubernetes.io/name: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: nginx-test
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/name: nginx
        helm.sh/chart: nginx-9.5.4
    spec:
      automountServiceAccountToken: false
      containers:
        - image: docker.io/bitnami/nginx:1.21.3-debian-10-r0
          imagePullPolicy: IfNotPresent
          name: nginx
      serviceAccountName: default
      shareProcessNamespace: false
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddAppLabels([]byte(tt.input), labels)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOutput, string(result))
		})
	}
}

func Test_DocumentSeperator(t *testing.T) {
	labels := KubeAppLabels{
		StackID:   123,
		StackName: "best-name",
		Owner:     "best-owner",
		Kind:      "git",
	}

	input := `apiVersion: v1
kind: Service
metadata:
  labels:
    io.kompose.service: database
---
apiVersion: v1
kind: Service
metadata:
  labels:
    io.kompose.service: backend
`
	expected := `apiVersion: v1
kind: Service
metadata:
  labels:
    io.kompose.service: database
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
---
apiVersion: v1
kind: Service
metadata:
  labels:
    io.kompose.service: backend
    io.portainer.kubernetes.application.kind: git
    io.portainer.kubernetes.application.name: best-name
    io.portainer.kubernetes.application.owner: best-owner
    io.portainer.kubernetes.application.stack: best-name
    io.portainer.kubernetes.application.stackid: "123"
`
	result, err := AddAppLabels([]byte(input), labels.ToMap())
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func Test_GetNamespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "valid namespace",
			input: `apiVersion: v1
kind: Namespace
metadata:
  namespace: test-namespace
`,
			want: "test-namespace",
		},
		{
			name: "invalid namespace",
			input: `apiVersion: v1
kind: Namespace
`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetNamespace([]byte(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func Test_ExtractDocuments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name: "multiple documents",
			input: `apiVersion: v1
kind: Namespace
---
apiVersion: v1
kind: Service
`,
			want: []string{`apiVersion: v1
kind: Namespace
`, `apiVersion: v1
kind: Service
`},
		},
		{
			name: "single document",
			input: `apiVersion: v1
kind: Namespace
`,
			want: []string{`apiVersion: v1
kind: Namespace
`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ExtractDocuments([]byte(tt.input), nil)
			assert.NoError(t, err)
			for i := range results {
				assert.Equal(t, tt.want[i], string(results[i]))
			}
		})
	}
}

func Test_AddAppEnvVars(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		vars       map[string]string
		wantOutput string
	}{
		{
			name: "update existing env",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: dependent-envars-demo
spec:
  containers:
    - name: dependent-envars-demo
      args:
        - while true; do echo -en '\n'; printf UNCHANGED_REFERENCE=$UNCHANGED_REFERENCE'\n'; printf SERVICE_ADDRESS=$SERVICE_ADDRESS'\n';printf ESCAPED_REFERENCE=$ESCAPED_REFERENCE'\n'; sleep 30; done;
      command:
        - sh
        - -c
      image: busybox:1.28
      env:
        - name: SERVICE_PORT
          value: "80"
        - name: SERVICE_IP
          value: "172.17.0.1"
        - name: UNCHANGED_REFERENCE
          value: "$(PROTOCOL)://$(SERVICE_IP):$(SERVICE_PORT)"
        - name: PROTOCOL
          value: "https"
        - name: SERVICE_ADDRESS
          value: "$(PROTOCOL)://$(SERVICE_IP):$(SERVICE_PORT)"
        - name: ESCAPED_REFERENCE
          value: "$$(PROTOCOL)://$(SERVICE_IP):$(SERVICE_PORT)"
`,
			vars: map[string]string{
				"SERVICE_PORT": "8080",
				"SERVICE_IP":   "192.168.1.1",
			},
			wantOutput: `apiVersion: v1
kind: Pod
metadata:
  name: dependent-envars-demo
spec:
  containers:
    - args:
        - while true; do echo -en '\n'; printf UNCHANGED_REFERENCE=$UNCHANGED_REFERENCE'\n'; printf SERVICE_ADDRESS=$SERVICE_ADDRESS'\n';printf ESCAPED_REFERENCE=$ESCAPED_REFERENCE'\n'; sleep 30; done;
      command:
        - sh
        - -c
      env:
        - name: SERVICE_PORT
          value: "8080"
        - name: SERVICE_IP
          value: 192.168.1.1
        - name: UNCHANGED_REFERENCE
          value: $(PROTOCOL)://$(SERVICE_IP):$(SERVICE_PORT)
        - name: PROTOCOL
          value: https
        - name: SERVICE_ADDRESS
          value: $(PROTOCOL)://$(SERVICE_IP):$(SERVICE_PORT)
        - name: ESCAPED_REFERENCE
          value: $$(PROTOCOL)://$(SERVICE_IP):$(SERVICE_PORT)
      image: busybox:1.28
      name: dependent-envars-demo
`,
		},
		{
			name: "try to add new env",
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
`,
			vars: map[string]string{
				"SERVICE_PORT": "8080",
				"SERVICE_IP":   "192.168.1.1",
			},
			wantOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - image: nginx:1.14.2
          name: nginx
          ports:
            - containerPort: 80
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UpdateContainerEnv([]byte(tt.input), tt.vars)
			if err != nil {
				log.Fatal(err)
			}

			assert.NoError(t, err)
			if !assert.Equal(t, tt.wantOutput, string(result)) {
				err = os.WriteFile("/tmp/failed-yaml.yaml", result, 0644)
				if err != nil {
					log.Fatal(err)
				}
			}
		})
	}
}

func Test_GetResourcesFromManifest(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput []KubeResource
		filters    []string
	}{
		{
			name: "simple deployment",
			input: `apiVersion: v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx-deployment
  namespace: default
`,
			wantOutput: []KubeResource{
				{
					Name:      "nginx-deployment",
					Kind:      "deployment",
					Namespace: "default",
				},
			},
			filters: []string{"deployment"},
		},
		{
			name: "simple daemonset",
			input: `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  namespace: kube-system
`,
			wantOutput: []KubeResource{
				{
					Name:      "fluentd-elasticsearch",
					Kind:      "daemonset",
					Namespace: "kube-system",
				},
			},
			filters: []string{"DaemonSet"},
		},
		{
			name: "multifile deployment",
			input: `apiVersion: v1
kind: Namespace
metadata:
  name: portainer
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: portainer-sa-clusteradmin
  namespace: portainer
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: portainer-crb-clusteradmin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: portainer-sa-clusteradmin
  namespace: portainer
---
apiVersion: v1
kind: Service
metadata:
  name: portainer
  namespace: portainer
  labels:
    io.portainer.kubernetes.application.stack: portainer
spec:
  type: LoadBalancer
  selector:
    app: app-portainer
  ports:
    - name: http
      protocol: TCP
      port: 9000
      targetPort: 9000
    - name: edge
      protocol: TCP
      port: 8000
      targetPort: 8000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: portainer
  namespace: portainer
  labels:
    io.portainer.kubernetes.application.stack: portainer    
spec:
  selector:
    matchLabels:
      app: app-portainer
  template:
    metadata:
      labels:
        app: app-portainer
    spec:
      serviceAccountName: portainer-sa-clusteradmin
      containers:
      - name: portainer
        image: portainerci/portainer:develop
        imagePullPolicy: Always
        ports:
        - containerPort: 9000
          protocol: TCP
        - containerPort: 8000
          protocol: TCP
`,
			wantOutput: []KubeResource{
				{
					Kind:      "namespace",
					Name:      "portainer",
					Namespace: "",
				},
				{
					Kind:      "serviceaccount",
					Name:      "portainer-sa-clusteradmin",
					Namespace: "portainer",
				},
				{
					Kind:      "clusterrolebinding",
					Name:      "portainer-crb-clusteradmin",
					Namespace: "",
				},
				{
					Kind:      "service",
					Name:      "portainer",
					Namespace: "portainer",
				},
				{
					Kind:      "deployment",
					Name:      "portainer",
					Namespace: "portainer",
				},
			},
			filters: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetResourcesFromManifest([]byte(tt.input), tt.filters)
			if err != nil {
				log.Fatal(err)
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOutput, result)
		})
	}
}
