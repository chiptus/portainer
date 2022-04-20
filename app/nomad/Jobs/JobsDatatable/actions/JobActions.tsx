import { useEnvironmentId } from 'Portainer/hooks/useEnvironmentId';
import { useMutation } from 'react-query';

import { Job } from '@/nomad/types';
import { confirmDeletionAsync } from '@/portainer/services/modal.service/confirm';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';

import { deleteJobs } from './delete';

interface Props {
  selectedItems: Job[];
  refreshData: () => Promise<void>;
}

export function JobActions({ selectedItems, refreshData }: Props) {
  const environmentId = useEnvironmentId();

  const mutation = useMutation(() => deleteJobs(environmentId, selectedItems));

  return (
    <LoadingButton
      loadingText="Removing..."
      isLoading={mutation.isLoading}
      disabled={selectedItems.length < 1 || mutation.isLoading}
      color="danger"
      onClick={handleDeleteClicked}
    >
      <i className="fa fa-trash-alt space-right" aria-hidden="true" />
      Remove
    </LoadingButton>
  );

  async function handleDeleteClicked() {
    const confirmed = await confirmDeletionAsync(
      'Are you sure to delete all selected jobs?'
    );

    if (!confirmed) {
      return;
    }

    mutation.mutate(undefined, {
      onSuccess() {
        return refreshData();
      },
    });
  }
}
