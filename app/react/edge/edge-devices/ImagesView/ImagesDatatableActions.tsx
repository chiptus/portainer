import { useRouter } from '@uirouter/react';
import { Trash2 } from 'lucide-react';

import * as notifications from '@/portainer/services/notifications';
import type { EnvironmentId } from '@/react/portainer/environments/types';
import { ImagesListResponse } from '@/react/docker/images/queries/useImages';

import { ButtonGroup, Button } from '@@/buttons';

import { removeImage } from './images.service';

interface Props {
  selectedItems: ImagesListResponse[];
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

  async function onRemoveClick(
    selectedItems: ImagesListResponse[],
    force?: boolean
  ) {
    const images = selectedItems;

    for (let i = 0; i < images.length; i += 1) {
      const image = images[i];
      try {
        await removeImage(endpointId, image.id, force);
        notifications.success(
          'Image removal successfully planned',
          image.tags[0]
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
