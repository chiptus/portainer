package client

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
)

type (
	// NomadClient represent a service used to execute Nomad operations
	NomadClient struct {
		nomadApiClient       *nomad.Client
		reverseTunnelService portaineree.ReverseTunnelService
		signatureService     portaineree.DigitalSignatureService
		lock                 *sync.Mutex
		endpoint             *portaineree.Endpoint
		tunnel               portaineree.TunnelDetails
	}
)

func NewClient(endpoint *portaineree.Endpoint, reverseTunnelService portaineree.ReverseTunnelService, signatureService portaineree.DigitalSignatureService) (portaineree.NomadClient, error) {
	nomadClientStr := NomadClient{
		reverseTunnelService: reverseTunnelService,
		signatureService:     signatureService,
		endpoint:             endpoint,
		lock:                 &sync.Mutex{},
	}

	err := nomadClientStr.initNomadApiClient()
	if err != nil {
		return nil, err
	}

	return &nomadClientStr, nil
}

func (c *NomadClient) initNomadApiClient() error {
	if c.endpoint.Type != portaineree.EdgeAgentOnNomadEnvironment {
		return errors.New("unsupported environment type")
	}

	var err error
	c.tunnel, err = c.reverseTunnelService.GetActiveTunnel(c.endpoint)
	if err != nil {
		return err
	}
	endpointURL := fmt.Sprintf("http://127.0.0.1:%d", c.tunnel.Port)

	signature, err := c.signatureService.CreateSignature(portaineree.PortainerAgentSignatureMessage)
	if err != nil {
		return err
	}

	httpClient := &http.Client{
		Transport: &NomadClientTransport{
			signatureHeader: signature,
			publicKeyHeader: c.signatureService.EncodedPublicKey(),
			tunnelAddress:   fmt.Sprintf("127.0.0.1:%d", c.tunnel.Port),
		},
	}

	config := &nomad.Config{
		Address:    endpointURL,
		HttpClient: httpClient,
	}

	c.nomadApiClient, err = nomad.NewClient(config)

	return err
}

func (c *NomadClient) setTunnelStatusToIdle(err error) {
	if strings.Contains(err.Error(), "connection refused") {
		c.reverseTunnelService.SetTunnelStatusToIdle(c.endpoint.ID)
	}
}

func (c *NomadClient) Validate() bool {
	tunnel := c.reverseTunnelService.GetTunnelDetails(c.endpoint.ID)
	return tunnel.Port == c.tunnel.Port && tunnel.Status == portaineree.EdgeAgentActive
}

func (c *NomadClient) Leader() (string, error) {
	leader, err := c.nomadApiClient.Status().Leader()
	if err != nil {
		c.setTunnelStatusToIdle(err)
	}
	return leader, err
}

func (c *NomadClient) ListJobs(namespace string) ([]*nomad.JobListStub, error) {
	jobList, _, err := c.nomadApiClient.Jobs().List(&nomad.QueryOptions{Namespace: namespace})
	if err != nil {
		c.setTunnelStatusToIdle(err)
	}
	return jobList, err
}

func (c *NomadClient) DeleteJob(jobID, namespace string) error {
	_, _, err := c.nomadApiClient.Jobs().Deregister(jobID, true, &nomad.WriteOptions{Namespace: namespace})
	if err != nil {
		c.setTunnelStatusToIdle(err)
	}
	return err
}

func (c *NomadClient) ListNodes() ([]*nomad.NodeListStub, error) {
	nodeList, _, err := c.nomadApiClient.Nodes().List(&nomad.QueryOptions{})
	if err != nil {
		c.setTunnelStatusToIdle(err)
	}
	return nodeList, err
}

func (c *NomadClient) ListAllocations(jobID, namespace string) ([]*nomad.AllocationListStub, error) {
	deployment, _, err := c.nomadApiClient.Jobs().LatestDeployment(jobID, &nomad.QueryOptions{Namespace: namespace})
	if err != nil {
		c.setTunnelStatusToIdle(err)
		return nil, err
	}

	if deployment == nil {
		return nil, fmt.Errorf("failed to get the latest deployment for job %s in namespace %s", jobID, namespace)
	}

	allocationsList, _, err := c.nomadApiClient.Deployments().Allocations(deployment.ID, &nomad.QueryOptions{})
	if err != nil {
		c.setTunnelStatusToIdle(err)
	}

	return allocationsList, err
}

// TaskEvents return all the Nomad task events belongs to the most recent deployment
func (c *NomadClient) TaskEvents(allocationID, taskName, namespace string) ([]*nomad.TaskEvent, error) {

	// retrieve allocation info
	allocation, _, err := c.nomadApiClient.Allocations().Info(allocationID, &nomad.QueryOptions{Namespace: namespace})
	if err != nil {
		c.setTunnelStatusToIdle(err)
		return nil, err
	}
	//retrieve task states via task name
	events := allocation.TaskStates[taskName].Events

	//sort events based on time desc order
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Time > events[j].Time
	})

	return events, nil
}

// TaskLogs return all the Nomad task logs
func (c *NomadClient) TaskLogs(refresh bool, allocationID, taskName, logType, origin, namespace string, offset int64) (<-chan *nomad.StreamFrame, error) {
	// retrieve allocation info
	allocation, _, err := c.nomadApiClient.Allocations().Info(allocationID, &nomad.QueryOptions{Namespace: namespace})
	if err != nil {
		c.setTunnelStatusToIdle(err)
		return nil, err
	}

	// retrieve logs stream as a channel
	logChan, _ := c.nomadApiClient.AllocFS().Logs(allocation, refresh, taskName, logType, origin, offset, nil, &nomad.QueryOptions{Namespace: namespace})

	return logChan, nil
}
