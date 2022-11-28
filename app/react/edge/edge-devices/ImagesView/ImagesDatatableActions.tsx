import { useRouter } from '@uirouter/react';
import { Trash2 } from 'lucide-react';

import * as notifications from '@/portainer/services/notifications';
import type { EnvironmentId } from '@/react/portainer/environments/types';
import { DockerImage } from '@/react/docker/images/types';

import { ButtonGroup, Button } from '@@/buttons';

import { removeImage } from './images.service';

interface Props {
  selectedItems: DockerImage[];
  endpointId: EnvironmentId;
}

export function ImagesDatatableActions({ selectedItems, endpointId }: Props) {
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

  async function onRemoveClick(selectedItems: DockerImage[], force?: boolean) {
    const images = selectedItems;

    for (let i = 0; i < images.length; i += 1) {
      const image = images[i];
      try {
        await removeImage(endpointId, image, force);
        notifications.success(
          'Image removal successfully planned',
          image.RepoTags[0]
        );
      } catch (err) {
        notifications.error(
          'Failure',
          err as Error,
          'Unable to schedule image removal'
        );
      }
    }

    router.stateService.reload();
  }
}
