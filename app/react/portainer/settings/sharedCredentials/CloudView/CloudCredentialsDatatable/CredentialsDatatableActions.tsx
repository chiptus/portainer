import { useRouter } from '@uirouter/react';
import { Trash2, Plus } from 'lucide-react';

import { confirmDelete } from '@@/modals/confirm';
import { Button } from '@@/buttons';
import { Icon } from '@@/Icon';

import { Credential } from '../../types';
import { useDeleteCredentialMutation } from '../../cloudSettings.service';

interface Props {
  selectedItems: Credential[];
}

export function CredentialsDatatableActions({ selectedItems }: Props) {
  const router = useRouter();
  const deleteCredentialMutation = useDeleteCredentialMutation();

  return (
    <>
      <Button
        disabled={selectedItems.length < 1}
        color="dangerlight"
        onClick={() => onDeleteClick(selectedItems)}
        data-cy="credentials-deleteButton"
      >
        <Icon icon={Trash2} className="space-right" aria-hidden="true" />
        Remove
      </Button>

      <Button
        onClick={() =>
          router.stateService.go(
            'portainer.settings.sharedcredentials.addCredential'
          )
        }
        data-cy="credentials-addButton"
      >
        <Icon icon={Plus} className="space-right" />
        Add credentials
      </Button>
    </>
  );

  async function onDeleteClick(selectedItems: Credential[]) {
    const confirmed = await confirmDelete(
      `Are you sure you want to remove the selected credentials?`
    );

    if (!confirmed) {
      return;
    }

    selectedItems.forEach((item) => {
      deleteCredentialMutation.mutate(item);
    });

    router.stateService.reload();
  }
}
