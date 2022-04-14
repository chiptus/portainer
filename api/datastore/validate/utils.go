package validate

import portaineree "github.com/portainer/portainer-ee/api"

func isRoleExists(roles []portaineree.Role, role portaineree.Role) bool {
	for _, r := range roles {
		if r.ID == role.ID {
			return true
		}
	}
	return false
}
