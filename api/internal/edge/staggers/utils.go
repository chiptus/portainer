package staggers

import (
	"errors"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
)

// buildStaggerQueueWithFixedDeviceNumber builds a stagger queue with fixed device number
// e.g. when the fixed device number is 2 and there are 7 endpoints
// related to the edge stack. It will result in [[1,2],[3,4],[5,6],[7]]
func buildStaggerQueueWithFixedDeviceNumber(arr []portaineree.EndpointID, deviceNumber int) [][]portaineree.EndpointID {
	length := len(arr) / deviceNumber
	if len(arr)%deviceNumber != 0 {
		length++
	}

	queue := make([][]portaineree.EndpointID, length)

	for i, endpointId := range arr {
		queue[i/deviceNumber] = append(queue[i/deviceNumber], endpointId)
	}

	return queue
}

func buildStaggerQueueWithIncrementalDeviceNumber(arr []portaineree.EndpointID, startFrom, incremental int) [][]portaineree.EndpointID {
	lowerBound := startFrom
	d := startFrom * incremental

	length := 1
	for i := 0; i < len(arr); i++ {
		if i >= lowerBound {
			lowerBound = lowerBound + d
			d *= incremental
			length++
		}
	}

	queue := make([][]portaineree.EndpointID, length)

	lowerBound = startFrom
	d = startFrom * incremental
	index := 0
	for i := 0; i < len(arr); i++ {
		if i >= lowerBound {
			lowerBound = lowerBound + d
			d *= incremental
			index++
		}

		queue[index] = append(queue[index], arr[i])
	}

	return queue
}

func calculateNextStaggerCheckIntervalForAsyncUpdate(edge *portaineree.EnvironmentEdgeSettings) int {
	arr := []int{edge.PingInterval, edge.SnapshotInterval, edge.SnapshotInterval}

	min := arr[0]
	for _, val := range arr {
		if val < min {
			min = val
		}
	}

	if min <= 0 {
		// interval default setting is -1
		return 30
	}
	return min / 2
}

func retry(fn func(i int) error, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn(i)
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return errors.New("failed after attempts")
}
