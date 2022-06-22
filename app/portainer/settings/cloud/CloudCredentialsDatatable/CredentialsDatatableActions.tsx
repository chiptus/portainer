import { useRouter } from '@uirouter/react';

import { confirmAsync } from '@/portainer/services/modal.service/confirm';

import { Button } from '@@/buttons';

import { Credential } from '../types';
import { useDeleteCredentialMutation } from '../cloudSettings.service';

interface Props {
  selectedItems: Credential[];
}

export function CredentialsDatatableActions({ selectedItems }: Props) {
  const router = useRouter();
  const deleteCredentialMutation = useDeleteCredentialMutation();

  return (
    <div className="actionBar">
      <Button
        disabled={selectedItems.length < 1}
        color="danger"
        onClick={() => onDeleteClick(selectedItems)}
        dataCy="credentials-deleteButton"
      >
        <i className="fa fa-trash-alt space-right" aria-hidden="true" />
        Remove
      </Button>

      <Button
        onClick={() =>
          router.stateService.go('portainer.settings.cloud.addCredential')
        }
        dataCy="credentials-addButton"
      >
        <i className="fa fa-plus space-right" aria-hidden="true" />
        Add credentials
      </Button>
    </div>
  );

  async function onDeleteClick(selectedItems: Credential[]) {
    const confirmed = await confirmAsync({
      title: 'Confirm action',
      message: `Are you sure you want to remove the selected ${
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
      deleteCredentialMutation.mutate(item);
    });
  }
}
