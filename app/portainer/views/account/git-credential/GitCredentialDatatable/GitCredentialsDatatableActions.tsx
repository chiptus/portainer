import { useRouter } from '@uirouter/react';
import { Plus, Trash2 } from 'lucide-react';

import { confirmDestructive } from '@@/modals/confirm';
import { Icon } from '@@/Icon';
import { Button } from '@@/buttons';

import { GitCredential } from '../types';
import { useDeleteGitCredentialMutation } from '../gitCredential.service';

interface Props {
  selectedItems: GitCredential[];
}

export function GitCredentialsDatatableActions({ selectedItems }: Props) {
  const router = useRouter();
  const deleteGitCredentialMutation = useDeleteGitCredentialMutation();

  return (
    <>
      <Button
        disabled={selectedItems.length < 1}
        color="dangerlight"
        onClick={() => onDeleteClick(selectedItems)}
        data-cy="credentials-deleteButton"
      >
        <Icon icon={Trash2} className="vertical-center" />
        Remove
      </Button>

      <Button
        onClick={() =>
          router.stateService.go('portainer.account.gitCreateGitCredential')
        }
        data-cy="credentials-addButton"
      >
        <Icon icon={Plus} className="vertical-center" />
        Add git credential
      </Button>
    </>
  );

  async function onDeleteClick(selectedItems: GitCredential[]) {
    const confirmed = await confirmDestructive({
      title: 'Confirm action',
      message: `Are you sure you want to remove the selected git ${
        selectedItems.length > 1 ? 'credentials' : 'credential'
      }?`,
    });

    if (!confirmed) {
      return;
    }

    selectedItems.forEach((item) => {
      deleteGitCredentialMutation.mutate(item);
    });
  }
}
