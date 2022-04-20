package snapshot

import (
	"log"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/nomad/clientFactory"
)

type Snapshotter struct {
	clientFactory *clientFactory.ClientFactory
}

// NewSnapshotter returns a new Snapshotter instance
func NewSnapshotter(clientFactory *clientFactory.ClientFactory) *Snapshotter {
	return &Snapshotter{
		clientFactory: clientFactory,
	}
}

// CreateSnapshot creates a snapshot of a specific Kubernetes environment(endpoint)
func (snapshotter *Snapshotter) CreateSnapshot(endpoint *portaineree.Endpoint) (*portaineree.NomadSnapshot, error) {
	client, err := snapshotter.clientFactory.GetClient(endpoint)
	if err != nil {
		return nil, err
	}

	return doSnapshot(client, endpoint)
}

func doSnapshot(client portaineree.NomadClient, endpoint *portaineree.Endpoint) (*portaineree.NomadSnapshot, error) {
	snapshot := &portaineree.NomadSnapshot{}

	// job count
	jobList, err := client.ListJobs("*")
	if err != nil {
		log.Printf("[WARN] [Nomad,snapshot] [message: unable to snapshot Nomad jobs] [endpoint: %s] [err: %s]", endpoint.Name, err)
		return nil, err
	}
	snapshot.JobCount = len(jobList)

	// group and task count
	for _, job := range jobList {
		groups := job.JobSummary.Summary
		snapshot.GroupCount += len(groups)

		for _, group := range groups {
			snapshot.TaskCount += group.Lost + group.Failed + group.Queued + group.Complete + group.Running + group.Starting
			snapshot.RunningTaskCount += group.Running
		}

		snapshotJob := portaineree.NomadSnapshotJob{
			ID:         job.ID,
			Status:     job.Status,
			Namespace:  job.Namespace,
			SubmitTime: time.UnixMicro(job.SubmitTime).Unix(),
			Tasks:      []portaineree.NomadSnapshotTask{},
		}

		allocations, err := client.ListAllocations(snapshotJob.ID, job.Namespace)
		if err != nil {
			log.Printf("[WARN] [Nomad,snapshot] [message: unable to snapshot Nomad jobs] [endpoint: %s] [err: %s]", endpoint.Name, err)
			return nil, err
		}

		for _, allocation := range allocations {
			for taskName, taskState := range allocation.TaskStates {
				task := portaineree.NomadSnapshotTask{
					JobID:        snapshotJob.ID,
					Namespace:    job.Namespace,
					TaskName:     taskName,
					State:        taskState.State,
					TaskGroup:    allocation.TaskGroup,
					AllocationID: allocation.ID,
					StartedAt:    taskState.StartedAt,
				}
				snapshotJob.Tasks = append(snapshotJob.Tasks, task)
			}
		}

		snapshot.Jobs = append(snapshot.Jobs, snapshotJob)
	}

	// node count
	nodeList, err := client.ListNodes()
	if err != nil {
		log.Printf("[WARN] [Nomad,snapshot] [message: unable to snapshot Nomad nodes] [endpoint: %s] [err: %s]", endpoint.Name, err)
		return nil, err
	}
	snapshot.NodeCount = len(nodeList)
	snapshot.Time = time.Now().Unix()

	return snapshot, nil
}
