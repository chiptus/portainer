import { notifySuccess } from '@/portainer/services/notifications';
import { useUpdateUserMutation } from '@/portainer/users/queries/useUpdateUserMutation';
import { Role } from '@/portainer/users/types';

import { DeleteButton } from '@@/buttons/DeleteButton';

import { TableEntry } from '../types';

export function TableActions({ selectedRows }: { selectedRows: TableEntry[] }) {
  const updateUserMutation = useUpdateUserMutation();

  return (
    <DeleteButton
      confirmMessage="You're about to remove edge administrators, doing so will remove their ability to manage all resources on all environments."
      disabled={selectedRows.length === 0}
      onConfirmed={() => handleRemove(selectedRows)}
    />
  );

  async function handleRemove(entries: TableEntry[]) {
    entries.forEach((e) => {
      if (e.type === 'user') {
        updateUserMutation.mutate(
          {
            userId: e.id,
            payload: { role: Role.Standard },
          },
          {
            onSuccess: () => {
              notifySuccess('Success', 'User successfully updated');
            },
          }
        );
      }
    });
  }
}
