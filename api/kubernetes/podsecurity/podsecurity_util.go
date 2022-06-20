package podsecurity

import (
	"context"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/client/conditions"
)

var (
	ErrPodNotFound       = errors.New("k8s/psppod not found")
	ErrContainerNotFound = errors.New("k8s/psppod/container not found")
)

// psppods allow k8s cluster to control security sensitive
// properties by applying policies
type psppod struct {
	clientSet kubernetes.Interface
	namespace string
	selector  string
}

func NewPSPPod(clientSet kubernetes.Interface) *psppod {
	return &psppod{
		clientSet: clientSet,
		namespace: GateKeeperNameSpace,
		selector:  GateKeeperSelector,
	}
}

func (p *psppod) waitFunc(ctx context.Context, podName string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := p.clientSet.CoreV1().Pods(p.namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, conditions.ErrPodCompleted
		}

		return false, nil
	}
}

func (p *psppod) pods(ctx context.Context) (*v1.PodList, error) {
	options := metav1.ListOptions{
		LabelSelector: p.selector,
	}

	return p.clientSet.CoreV1().Pods(p.namespace).List(ctx, options)
}

func (p *psppod) wait(ctx context.Context, interval, timeout time.Duration) error {
	podList, err := p.pods(ctx)
	if err != nil {
		return errors.Wrap(err, "fetch k8s/psppod error")
	}

	if len(podList.Items) == 0 {
		return ErrPodNotFound
	}

	for _, pod := range podList.Items {
		log.Debugf("waiting for k8s/psppod [name=%s] running", pod.Name)
		if err := wait.PollImmediate(interval, timeout, p.waitFunc(ctx, pod.Name)); err != nil {
			return errors.Wrap(err, "k8s/psppod running error")
		}
	}

	return nil
}

func (p *psppod) images(ctx context.Context) (map[string]string, error) {
	podList, err := p.pods(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch k8s/psppod error")
	}

	if len(podList.Items) == 0 {
		return nil, ErrPodNotFound
	}

	podsImage := map[string]string{}
	for _, item := range podList.Items {
		pod, err := p.clientSet.CoreV1().Pods(p.namespace).Get(ctx, item.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if len(pod.Status.ContainerStatuses) == 0 {
			return nil, ErrContainerNotFound
		}

		podsImage[item.Name] = pod.Status.ContainerStatuses[0].Image
	}

	return podsImage, nil
}

// WaitForOpaReady will wait until duration d (from now) for a gatekeeper deployment pod to reach defined phase/status.
// The pod status will be polled at specified delay until the pod reaches ready state.
func WaitForOpaReady(ctx context.Context, clientSet *kubernetes.Clientset) error {
	return NewPSPPod(clientSet).wait(ctx, GateKeeperInterval, GateKeeperTimeOut)
}
