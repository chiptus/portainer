import { confirmDestructive } from '@@/modals/confirm';
import { buildConfirmButton } from '@@/modals/utils';

export async function confirmStackDeletion() {
  const result = await confirmDestructive({
    title: 'Are you sure?',
    confirmButton: buildConfirmButton('Remove', 'danger'),
    message:
      'Do you want to remove the selected stack(s)? Associated services will be removed as well.',
  });

  return result;
}
