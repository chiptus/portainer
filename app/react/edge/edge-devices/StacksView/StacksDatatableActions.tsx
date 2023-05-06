import { useRouter } from '@uirouter/react';
import { Trash2 } from 'lucide-react';
import { useMutation } from 'react-query';

import * as notifications from '@/portainer/services/notifications';
import { confirmStackDeletion } from '@/react/docker/containers/common/confirm-stack-delete-modal';
import type { EnvironmentId } from '@/react/portainer/environments/types';

import { Button } from '@@/buttons';

import { removeStack } from './stacks.service';
import { StackInAsyncSnapshot } from './types';

interface Props {
  selectedItems: StackInAsyncSnapshot[];
  endpointId: EnvironmentId;
}

export function StacksDatatableActions({ selectedItems, endpointId }: Props) {
  const removeStackMutation = useMutation((stackId: number) =>
    removeStack(endpointId, stackId)
  );

  const router = useRouter();

  const selectedItemCount = selectedItems.length;
  async function onRemoveClick(selectedItems: StackInAsyncSnapshot[]) {
    const result = await confirmStackDeletion();
    if (!result) {
      return;
    }

    removeSelectedStacks(selectedItems);
  }

  async function removeSelectedStacks(stacks: StackInAsyncSnapshot[]) {
    for (let i = 0; i < stacks.length; i += 1) {
      const stack = stacks[i];

      if (stack.Metadata.stackId) {
        try {
          await removeStackMutation.mutateAsync(+stack.Metadata.stackId);
          notifications.success(
            'Stack removal successfully planned',
            stack.StackName ? stack.StackName : stack.Names[0]
          );
        } catch (err) {
          notifications.error(
            'Failure',
            err as Error,
            'Unable to schedule stack removal'
          );
        }
      }
    }

    router.stateService.reload();
  }

  return (
    <Button
      color="dangerlight"
      onClick={() => onRemoveClick(selectedItems)}
      disabled={selectedItemCount === 0 || removeStackMutation.isLoading}
      icon={Trash2}
    >
      Remove
    </Button>
  );
}
