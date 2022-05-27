package cli

import (
	"context"

	portaineree "github.com/portainer/portainer-ee/api"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // register GCP auth provider
)

var DefaultAgentVersion = portaineree.APIVersion

// GetPortainerAgentIP checks whether there is an IP address associated to the agent service and returns it.
func (kcl *KubeClient) GetPortainerAgentIPOrHostname() (string, error) {
	service, err := kcl.cli.CoreV1().Services("portainer").Get(context.TODO(), "portainer-agent", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		if service.Status.LoadBalancer.Ingress[0].IP != "" {
			return service.Status.LoadBalancer.Ingress[0].IP, nil
		}
		return service.Status.LoadBalancer.Ingress[0].Hostname, nil
	}

	return "", nil
}

// DeployPortainerAgent deploys the Portainer agent in the current Kubernetes
// environment it is effectively the equivalent of
// https://github.com/portainer/k8s/blob/master/deploy/manifests/agent/portainer-agent-k8s-lb.yaml
//
// This approach means we have another area to maintain, but allows KaaS
// provisioning without a public network interface. For example, you could in
// theory create several clusters on Linode, add them all to a private network,
// and still manage them with portainer even if they have heavily filtered
// public internet access (or even none at all).
func (kcl *KubeClient) DeployPortainerAgent() error {
	// NAMESPACE
	namespaceName := "portainer"

	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	_, err := kcl.cli.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	// SERVICE ACCOUNT
	serviceAccountName := "portainer-sa-clusteradmin"

	err = kcl.createUserServiceAccount(namespaceName, serviceAccountName)
	if err != nil {
		return err
	}

	// CLUSTER ROLE BINDING
	clusterRoleBindingName := "portainer-crb-clusteradmin"

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespaceName,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
	}

	_, err = kcl.cli.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	// SERVICE
	serviceName := "portainer-agent"

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				"app": "portainer-agent",
			},
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
					Port:       9001,
					TargetPort: intstr.FromInt(9001),
				},
			},
		},
	}

	_, err = kcl.cli.CoreV1().Services(namespaceName).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	// HEADLESS SERVICE
	headlessServiceName := "portainer-agent-headless"

	headlessService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: headlessServiceName,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app": "portainer-agent",
			},
		},
	}

	_, err = kcl.cli.CoreV1().Services(namespaceName).Create(context.TODO(), headlessService, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	// DEPLOYMENT
	deploymentName := "portainer-agent"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "portainer-agent",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "portainer-agent",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []v1.Container{
						{
							Name:            "portainer-agent",
							Image:           "portainer/agent:" + DefaultAgentVersion,
							ImagePullPolicy: v1.PullAlways,
							Env: []v1.EnvVar{
								{
									Name:  "LOG_LEVEL",
									Value: "INFO",
								},
								{
									Name:  "AGENT_CLUSTER_ADDR",
									Value: "portainer-agent-headless",
								},
								{
									Name: "KUBERNETES_POD_IP",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 9001,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = kcl.cli.AppsV1().Deployments(namespaceName).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
