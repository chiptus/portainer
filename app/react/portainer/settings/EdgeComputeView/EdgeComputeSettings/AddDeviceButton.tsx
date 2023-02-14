import { useRouter } from '@uirouter/react';
import { Plus } from 'lucide-react';
import { useState } from 'react';

import { Button } from '@@/buttons';
import { Dialog } from '@@/modals/Dialog';
import { buildCancelButton, buildConfirmButton } from '@@/modals/utils';
import { OnSubmit, openModal } from '@@/modals';

import { usePublicSettings } from '../../queries';

enum DeployType {
  FDO = 'FDO',
  MANUAL = 'MANUAL',
}

export function AddDeviceButton() {
  const router = useRouter();
  const isFDOEnabledQuery = usePublicSettings({
    select: (settings) => settings.IsFDOEnabled,
  });
  const isFDOEnabled = !!isFDOEnabledQuery.data;

  return (
    <Button onClick={handleNewDeviceClick} icon={Plus}>
      Add Device
    </Button>
  );

  async function handleNewDeviceClick() {
    const result = await getDeployType();

    switch (result) {
      case DeployType.FDO:
        router.stateService.go('portainer.endpoints.importDevice');
        break;
      case DeployType.MANUAL:
        router.stateService.go('portainer.wizard.endpoints', {
          edgeDevice: true,
        });
        break;
      default:
        break;
    }
  }

  function getDeployType() {
    if (!isFDOEnabled) {
      return Promise.resolve(DeployType.MANUAL);
    }

    return askForDeployType();
  }
}

function askForDeployType() {
  return openModal(AddDeviceDialog, {});
}

function AddDeviceDialog({ onSubmit }: { onSubmit: OnSubmit<DeployType> }) {
  const [deployType, setDeployType] = useState<DeployType>(DeployType.FDO);
  return (
    <Dialog
      title="How would you like to add an Edge Device?"
      message={
        <>
          <RadioInput
            name="deployType"
            value={DeployType.FDO}
            label="Provision bare-metal using Intel FDO"
            groupValue={deployType}
            onChange={setDeployType}
          />

          <RadioInput
            name="deployType"
            value={DeployType.MANUAL}
            onChange={setDeployType}
            groupValue={deployType}
            label="Deploy agent manually"
          />
        </>
      }
      buttons={[buildCancelButton(), buildConfirmButton()]}
      onSubmit={(confirm) => onSubmit(confirm ? deployType : undefined)}
    />
  );
}

function RadioInput<T extends number | string>({
  value,
  onChange,
  label,
  groupValue,
  name,
}: {
  value: T;
  onChange: (value: T) => void;
  label: string;
  groupValue: T;
  name: string;
}) {
  return (
    <label className="flex items-center gap-2">
      <input
        className="!m-0"
        type="radio"
        name={name}
        value={value}
        checked={groupValue === value}
        onChange={() => onChange(value)}
      />
      {label}
    </label>
  );
}
