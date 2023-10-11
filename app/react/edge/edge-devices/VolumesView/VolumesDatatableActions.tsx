import { useRouter } from '@uirouter/react';
import { Trash2 } from 'lucide-react';

import * as notifications from '@/portainer/services/notifications';
import type { EnvironmentId } from '@/react/portainer/environments/types';
import { VolumeViewModel } from '@/docker/models/volume';

import { ButtonGroup, Button } from '@@/buttons';

import { removeVolume } from './volumes.service';

interface Props {
  selectedItems: VolumeViewModel[];
  endpointId: EnvironmentId;
}

export function VolumesDatatableActions({ selectedItems, endpointId }: Props) {
  const selectedItemCount = selectedItems.length;

  const router = useRouter();

  return (
    <ButtonGroup>
      <Button
        color="dangerlight"
        onClick={() => onRemoveClick(selectedItems)}
        disabled={selectedItemCount === 0}
        icon={Trash2}
      >
        Remove
      </Button>
    </ButtonGroup>
  );

  async function onRemoveClick(selectedItems: VolumeViewModel[]) {
    const volumes = selectedItems;

    for (let i = 0; i < volumes.length; i += 1) {
      const volume = volumes[i];
      try {
        await removeVolume(endpointId, volume.Id);
        notifications.success('Volume removal successfully planned', volume.Id);
      } catch (err) {
        notifications.error(
          'Failure',
          err as Error,
          'Unable to schedule volume removal'
        );
      }
    }

    router.stateService.reload();
  }
}
