package cli

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kcl *KubeClient) NamespaceAccessPoliciesDeleteNamespace(ns string) error {
	kcl.lock.Lock()
	defer kcl.lock.Unlock()

	policies, err := kcl.GetNamespaceAccessPolicies()
	if err != nil {
		return errors.WithMessage(err, "failed to fetch access policies")
	}

	delete(policies, ns)

	return kcl.UpdateNamespaceAccessPolicies(policies)
}

// GetNamespaceAccessPolicies gets the namespace access policies
// from config maps in the portainer namespace
func (kcl *KubeClient) GetNamespaceAccessPolicies() (map[string]portaineree.K8sNamespaceAccessPolicy, error) {
	configMap, err := kcl.cli.CoreV1().ConfigMaps(portainerNamespace).Get(context.TODO(), portainerConfigMapName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	accessData := configMap.Data[portainerConfigMapAccessPoliciesKey]

	var policies map[string]portaineree.K8sNamespaceAccessPolicy
	err = json.Unmarshal([]byte(accessData), &policies)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

// UpdateNamespaceAccessPolicies updates the namespace access policies
func (kcl *KubeClient) UpdateNamespaceAccessPolicies(accessPolicies map[string]portaineree.K8sNamespaceAccessPolicy) error {
	data, err := json.Marshal(accessPolicies)
	if err != nil {
		return err
	}

	configMap, err := kcl.cli.CoreV1().ConfigMaps(portainerNamespace).Get(context.TODO(), portainerConfigMapName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	configMap.Data[portainerConfigMapAccessPoliciesKey] = string(data)
	_, err = kcl.cli.CoreV1().ConfigMaps(portainerNamespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
