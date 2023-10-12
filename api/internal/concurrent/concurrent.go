// Package concurrent provides utilities for running multiple functions concurrently in Go.
// For example, many kubernetes calls can take a while to fulfill.  Oftentimes in Portainer
// we need to get a list of objects from multiple kubernetes REST APIs. We can often call these
// apis concurrently to speed up the response time.
// This package provides a clean way to do just that.
//
// Examples:
//   The ConfigMaps and Secrets function converted using concurrent.Run.
/*

// GetConfigMapsAndSecrets gets all the ConfigMaps AND all the Secrets for a
// given namespace in a k8s endpoint. The result is a list of both config maps
// and secrets. The IsSecret boolean property indicates if a given struct is a
// secret or configmap.
func (kcl *KubeClient) GetConfigMapsAndSecrets(namespace string) ([]models.K8sConfigMapOrSecret, error) {

	// use closures to capture the current kube client and namespace by declaring wrapper functions
	// that match the interface signature for concurrent.Func

	listConfigMaps := func(ctx context.Context) (interface{}, error) {
		return kcl.cli.CoreV1().ConfigMaps(namespace).List(context.Background(), meta.ListOptions{})
	}

	listSecrets := func(ctx context.Context) (interface{}, error) {
		return kcl.cli.CoreV1().Secrets(namespace).List(context.Background(), meta.ListOptions{})
	}

	// run the functions concurrently and wait for results.  We can also pass in a context to cancel.
	// e.g. Deadline timer.
	results, err := concurrent.Run(context.TODO(), listConfigMaps, listSecrets)
	if err != nil {
		return nil, err
	}

	var configMapList *core.ConfigMapList
	var secretList *core.SecretList
	for _, r := range results {
		switch v := r.Result.(type) {
		case *core.ConfigMapList:
			configMapList = v
		case *core.SecretList:
			secretList = v
		}
	}

	// TODO: Applications
	var combined []models.K8sConfigMapOrSecret
	for _, m := range configMapList.Items {
		var cm models.K8sConfigMapOrSecret
		cm.UID = string(m.UID)
		cm.Name = m.Name
		cm.Namespace = m.Namespace
		cm.Annotations = m.Annotations
		cm.Data = m.Data
		cm.CreationDate = m.CreationTimestamp.Time.UTC().Format(time.RFC3339)
		combined = append(combined, cm)
	}

	for _, s := range secretList.Items {
		var secret models.K8sConfigMapOrSecret
		secret.UID = string(s.UID)
		secret.Name = s.Name
		secret.Namespace = s.Namespace
		secret.Annotations = s.Annotations
		secret.Data = msbToMss(s.Data)
		secret.CreationDate = s.CreationTimestamp.Time.UTC().Format(time.RFC3339)
		secret.IsSecret = true
		secret.SecretType = string(s.Type)
		combined = append(combined, secret)
	}

	return combined, nil
}

*/

package concurrent

import (
	"context"
	"sync"
)

// Result contains the result and any error returned from running a client task function
type Result struct {
	Result interface{} // the result of running the task function
	Err    error       // any error that occurred while running the task function
}

// Func is a function returns a result or error
type Func func(ctx context.Context) (interface{}, error)

// Run runs a list of functions returns the results
func Run(ctx context.Context, tasks ...Func) ([]Result, error) {
	var wg sync.WaitGroup
	resultsChan := make(chan Result, len(tasks))

	localCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	// run each task function in a separate goroutine
	for _, fn := range tasks {
		wg.Add(1)
		go func(fn Func) {
			defer wg.Done()
			result, err := fn(localCtx)
			resultsChan <- Result{Result: result, Err: err}
		}(fn)
	}

	// wait for all the goroutines to complete and close the results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// collect the results from the results channel. Cancel outstanding requests on failure
	results := make([]Result, 0, len(tasks))
	for r := range resultsChan {
		if r.Err != nil {
			cancelCtx()
			return nil, r.Err
		}

		results = append(results, r)
	}

	return results, nil
}
