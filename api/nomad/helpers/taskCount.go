package helpers

import "github.com/hashicorp/nomad/api"

func CalcGroupTasks(group api.TaskGroupSummary) int {
	return group.Lost + group.Failed + group.Queued + group.Complete + group.Running + group.Starting
}
