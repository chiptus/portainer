import axios, { parseAxiosError } from '@/portainer/services/axios';
import { TeamMembership } from '@/react/portainer/users/teams/types';

import { User, UserId } from './types';
import { filterNonAdministratorUsers } from './user.helpers';
import { buildUrl } from './buildUrl';

export async function getUsers(
  includeAdministrators = false,
  environmentId = 0
) {
  try {
    const { data } = await axios.get<User[]>(buildUrl(), {
      params: { environmentId },
    });

    return includeAdministrators ? data : filterNonAdministratorUsers(data);
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to retrieve users');
  }
}

export async function getUserMemberships(id: UserId) {
  try {
    const { data } = await axios.get<TeamMembership[]>(
      buildUrl(id, 'memberships')
    );
    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Unable to retrieve user memberships');
  }
}
