import { useState } from 'react';

import { pluralize } from '@/portainer/helpers/strings';

import { Modal, OnSubmit, ModalType, openModal } from '@@/modals';
import { Button } from '@@/buttons';
import { SwitchField } from '@@/form-components/SwitchField';
import { Input } from '@@/form-components/Input';

import { Environment } from '../types';

interface Props {
  onSubmit: OnSubmit<{ deleteClusters: boolean; confirmed: boolean }>;
  environmentsToDelete: Environment[];
}

function ConfirmDeleteEnvironmentsModal({
  onSubmit,
  environmentsToDelete,
}: Props) {
  // force delete is only available for deletable clusters (only microk8s currently)
  // force delete enabled will delete the cluster and remove it from portainer.
  // force delete disabled will remove the cluster from portainer but not delete it.
  const [deleteClusters, setDeleteClusters] = useState(false);
  const [confirmText, setConfirmText] = useState('');

  const microk8sClusters = environmentsToDelete.filter(
    (env) => env.CloudProvider?.Provider === 'microk8s'
  );

  return (
    <Modal
      onDismiss={() => onSubmit()}
      aria-label="confirm delete environment modal"
    >
      <Modal.Header title="Are you sure?" modalType={ModalType.Destructive} />

      <Modal.Body>
        <p>
          Are you sure you want to disconnect{' '}
          <b>{environmentsToDelete.length}</b>{' '}
          {pluralize(environmentsToDelete.length, 'environment')} and remove
          their configurations from Portainer?
        </p>
        {!!microk8sClusters.length && (
          <SwitchField
            name="pullLatest"
            label={`Also permanently delete ${microk8sClusters.length} cluster(s) and uninstall MicroK8s from them?`}
            checked={deleteClusters}
            onChange={(val) => {
              setDeleteClusters(val);
              setConfirmText('');
            }}
          />
        )}
        {deleteClusters && (
          <div className="flex flex-col">
            <ul className="mt-2 max-h-96 list-inside overflow-hidden overflow-y-auto text-sm">
              {microk8sClusters.map((env) => (
                <li key={env.Id}>{env.Name}</li>
              ))}
            </ul>
            <p>
              To proceed, type{' '}
              <span className="text-error-7 th-highcontrast:text-error-6 th-dark:text-error-6">
                DELETE
              </span>{' '}
              in the field below.
            </p>
            <Input
              name="confirm"
              type="text"
              placeholder="Type DELETE to confirm"
              className="w-full"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
            />
          </div>
        )}
      </Modal.Body>
      <Modal.Footer>
        <Button
          onClick={() => onSubmit({ deleteClusters, confirmed: false })}
          color="default"
        >
          Cancel
        </Button>
        <Button
          onClick={() => onSubmit({ deleteClusters, confirmed: true })}
          color="danger"
          disabled={deleteClusters && confirmText !== 'DELETE'}
        >
          Confirm
        </Button>
      </Modal.Footer>
    </Modal>
  );
}

export async function confirmDeleteEnvironments(
  environmentsToDelete: Environment[]
) {
  return openModal(ConfirmDeleteEnvironmentsModal, {
    environmentsToDelete,
  });
}
