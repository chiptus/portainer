import { Trash2 } from 'lucide-react';

import { notifySuccess } from '@/portainer/services/notifications';
import { useDeleteEnvironmentsMutation } from '@/react/portainer/environments/queries/useDeleteEnvironmentsMutation';
import { Environment } from '@/react/portainer/environments/types';

import { Button } from '@@/buttons';
import { ModalType } from '@@/modals';
import { confirm } from '@@/modals/confirm';
import { buildConfirmButton } from '@@/modals/utils';
import { TooltipWithChildren } from '@@/Tip/TooltipWithChildren';

import { useAssociateDeviceMutation, useLicenseOverused } from '../queries';

export function TableActions({
  selectedRows,
}: {
  selectedRows: Environment[];
}) {
  const associateMutation = useAssociateDeviceMutation();
  const removeMutation = useDeleteEnvironmentsMutation();
  const licenseOverused = useLicenseOverused(selectedRows.length);

  return (
    <>
      <Button
        onClick={() => handleRemoveDevice(selectedRows)}
        disabled={selectedRows.length === 0}
        color="dangerlight"
        icon={Trash2}
      >
        Remove Device
      </Button>

      <TooltipWithChildren
        message={
          licenseOverused && (
            <>
              Associating devices is disabled as your node count exceeds your
              license limit
            </>
          )
        }
      >
        <span>
          <Button
            onClick={() => handleAssociateDevice(selectedRows)}
            disabled={selectedRows.length === 0 || licenseOverused}
          >
            Associate Device
          </Button>
        </span>
      </TooltipWithChildren>
    </>
  );

  function handleAssociateDevice(devices: Environment[]) {
    associateMutation.mutate(
      devices.map((d) => d.Id),
      {
        onSuccess() {
          notifySuccess('Success', 'Edge devices associated successfully');
        },
      }
    );
  }

  async function handleRemoveDevice(devices: Environment[]) {
    const confirmed = await confirm({
      title: 'Are you sure?',
      message:
        "You're about to remove edge device(s) from waiting room, which will not be shown until next agent startup.",
      confirmButton: buildConfirmButton('Remove', 'danger'),
      modalType: ModalType.Destructive,
    });

    if (!confirmed) {
      return;
    }

    removeMutation.mutate(
      devices.map((d) => d.Id),
      {
        onSuccess() {
          notifySuccess('Success', 'Edge devices were hidden successfully');
        },
      }
    );
  }
}
