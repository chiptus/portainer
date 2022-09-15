import { useRouter } from '@uirouter/react';

import { confirmDestructiveAsync } from '@/portainer/services/modal.service/confirm';

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
        color="danger"
        onClick={() => onDeleteClick(selectedItems)}
        data-cy="credentials-deleteButton"
      >
        <i className="fa fa-trash-alt space-right" aria-hidden="true" />
        Remove
      </Button>

      <Button
        onClick={() =>
          router.stateService.go('portainer.account.gitCreateGitCredential')
        }
        data-cy="credentials-addButton"
      >
        <i className="fa fa-plus space-right" aria-hidden="true" />
        Add git credential
      </Button>
    </>
  );

  async function onDeleteClick(selectedItems: GitCredential[]) {
    const confirmed = await confirmDestructiveAsync({
      title: 'Confirm action',
      message: `Are you sure you want to remove the selected git ${
        selectedItems.length > 1 ? 'credentials' : 'credential'
      }?`,
      buttons: {
        cancel: {
          label: 'Cancel',
          className: 'btn-default',
        },
        confirm: {
          label: 'Confirm',
          className: 'btn-primary',
        },
      },
    });

    if (!confirmed) {
      return;
    }

    selectedItems.forEach((item) => {
      deleteGitCredentialMutation.mutate(item);
    });
  }
}
