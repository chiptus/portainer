package staggers

import (
	"errors"
	"regexp"

	portaineree "github.com/portainer/portainer-ee/api"
)

func ValidateStaggerConfig(config *portaineree.EdgeStaggerConfig) error {
	if config != nil && config.StaggerOption != portaineree.EdgeStaggerOptionAllAtOnce {
		if config.StaggerOption != portaineree.EdgeStaggerOptionParallel {
			return errors.New("invalid stagger option")
		}

		if config.StaggerParallelOption != portaineree.EdgeStaggerParallelOptionFixed &&
			config.StaggerParallelOption != portaineree.EdgeStaggerParallelOptionIncremental {
			return errors.New("invalid stagger parallel option")
		}

		if config.StaggerParallelOption == portaineree.EdgeStaggerParallelOptionFixed &&
			config.DeviceNumber == 0 {
			return errors.New("invalid device number")
		}

		if config.StaggerParallelOption == portaineree.EdgeStaggerParallelOptionIncremental &&
			config.DeviceNumberStartFrom == 0 {
			return errors.New("invalid device number start from")
		}

		if config.UpdateFailureAction != portaineree.EdgeUpdateFailureActionContinue &&
			config.UpdateFailureAction != portaineree.EdgeUpdateFailureActionPause &&
			config.UpdateFailureAction != portaineree.EdgeUpdateFailureActionRollback {
			return errors.New("invalid update failure action")
		}

		if config.Timeout != "" && config.Timeout != "0" {
			regex := regexp.MustCompile(`^[1-9][0-9]*$`)
			if !regex.MatchString(config.Timeout) {
				return errors.New("invalid timeout")
			}
		}

		if config.UpdateDelay != "" && config.UpdateDelay != "0" {
			regex := regexp.MustCompile(`^[1-9][0-9]*$`)
			if !regex.MatchString(config.UpdateDelay) {
				return errors.New("invalid update delay")
			}
		}
	}

	return nil
}
