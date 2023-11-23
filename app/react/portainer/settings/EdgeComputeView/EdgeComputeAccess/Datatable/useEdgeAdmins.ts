import { useMemo } from 'react';

import { useUsers as useUsersList } from '@/portainer/users/queries';
import { Role } from '@/portainer/users/types';

import { TableEntry } from '../types';

export function useEdgeAdmins() {
  const usersQuery = useUsersList(false, 0, true, (users) =>
    users.filter((u) => u.Role === Role.EdgeAdmin)
  );

  const data: TableEntry[] = useMemo(
    () => [
      ...(!usersQuery.data
        ? []
        : usersQuery.data.map(
            (u): TableEntry => ({
              id: u.Id,
              name: u.Username,
              type: 'user',
            })
          )),
    ],
    [usersQuery.data]
  );

  return useMemo(
    () => ({ data, isLoading: usersQuery.isLoading }),
    [data, usersQuery.isLoading]
  );
}
