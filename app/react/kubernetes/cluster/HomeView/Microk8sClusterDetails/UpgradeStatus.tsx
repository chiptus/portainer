import { useEffect, useState } from 'react';
import { Loader2, X } from 'lucide-react';
import clsx from 'clsx';
import sanitize from 'sanitize-html';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { useUpdateEnvironmentMutation } from '@/react/portainer/environments/queries/useUpdateEnvironmentMutation';
import { OperationStatus } from '@/react/portainer/environments/types';
import { queryClient } from '@/react-tools/react-query';
import { environmentQueryKeys } from '@/react/portainer/environments/queries/query-keys';

import { AlertContainer, alertSettings } from '@@/Alert/Alert';
import { Icon } from '@@/Icon';
import { TextTip } from '@@/Tip/TextTip';
import { Button } from '@@/buttons';
import { confirmDestructive } from '@@/modals/confirm';
import { buildConfirmButton } from '@@/modals/utils';

export function UpgradeStatus() {
  const [autoRefreshRate, setAutoRefreshRate] = useState<number | undefined>();
  const [currentOperationStatus, setCurrentOperationStatus] = useState<
    OperationStatus | undefined
  >('');
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId, (env) => env, {
    autoRefreshRate,
  });
  const operationStatus = environment?.StatusMessage?.operationStatus;
  const updateEnvironmentStatusMutation = useUpdateEnvironmentMutation();

  useEffect(() => {
    // if operationStatus is processing, change the autorefresh rate to 10000ms, otherwise set it to undefined
    if (operationStatus === 'processing') {
      setAutoRefreshRate(3000);
    } else {
      setAutoRefreshRate(undefined);
    }
    // when the operation finished processing, notify the user, and invalidate the cluster queries to force a refresh
    if (
      currentOperationStatus === 'processing' &&
      operationStatus !== 'processing'
    ) {
      // all query keys I want to invalidate have ['envirnments', environmentId] as their first two elements
      setTimeout(
        () =>
          queryClient.invalidateQueries(
            environmentQueryKeys.item(environmentId)
          ),
        2000 // there's a delay to show updated nodes from scaling, so wait before calling invalidateQueries
      );
    }
    setCurrentOperationStatus(operationStatus);
  }, [currentOperationStatus, environmentId, operationStatus]);

  // '' is an idle status
  if (operationStatus === '') {
    return null;
  }

  return (
    <AlertContainer
      className={clsx(
        operationStatus === 'warning' && alertSettings.warn.container,
        operationStatus === 'error' && alertSettings.error.container,
        operationStatus !== 'error' &&
          operationStatus !== 'warning' &&
          alertSettings.info.container
      )}
    >
      <div className="flex flex-col gap-y-4 text-sm">
        <div className={clsx('flex items-center justify-between')}>
          <div
            className={clsx(
              'flex items-center',
              operationStatus === 'error' && alertSettings.error.body,
              operationStatus === 'warning' && alertSettings.warn.body,
              operationStatus !== 'error' &&
                operationStatus !== 'warning' &&
                alertSettings.info.body
            )}
          >
            {operationStatus === 'processing' && (
              <Icon icon={Loader2} className="!mr-2 animate-spin-slow" />
            )}
            {environment?.StatusMessage?.summary}
          </div>
          {operationStatus === 'warning' && (
            <Button
              icon={X}
              className={clsx(
                'flex !text-gray-7 hover:!text-gray-8 th-highcontrast:!text-gray-6 th-highcontrast:hover:!text-gray-5 th-dark:!text-gray-6 th-dark:hover:!text-gray-5'
              )}
              color="link"
              size="small"
              onClick={() => dismissMessage()}
              title="Dismiss message"
            />
          )}
          {operationStatus === 'processing' && (
            <Button
              icon={X}
              className={clsx('flex')}
              color="dangerlight"
              size="small"
              onClick={() => {
                confirmDestructive({
                  title: 'Are you sure?',
                  message:
                    'You should only clear the status if any processing of your cluster has hung (e.g. processing of an upgrade or an enable/disable of an addon). This will not resolve any issues in the cluster, so, please then use the Status option on your nodes to determine their condition. Are you sure you want to clear the status?',
                  cancelButtonLabel: 'Cancel',
                  confirmButton: buildConfirmButton('Clear status', 'danger'),
                })
                  .then((confirmed) => {
                    if (confirmed) {
                      dismissMessage();
                    }
                    return true;
                  })
                  .catch(() => {});
              }}
              title="Dismiss message"
            >
              Clear &apos;processing&apos; status
            </Button>
          )}
        </div>
        {environment?.StatusMessage?.detail && (
          <p
            className="mb-0 text-xs" // eslint-disable-next-line react/no-danger
            dangerouslySetInnerHTML={{
              __html: sanitize(
                `<div>${environment.StatusMessage.detail}</div>`
              ),
            }}
          />
        )}
        {operationStatus === 'processing' && (
          <TextTip color="blue">
            Further cluster management operations (upgrading, enabling/disabling
            of addons, adding/removing of nodes) cannot be initiated whilst
            cluster changes are being processed.
          </TextTip>
        )}
      </div>
    </AlertContainer>
  );

  function dismissMessage() {
    updateEnvironmentStatusMutation.mutate({
      id: environmentId,
      payload: {
        IsSetStatusMessage: true,
        StatusMessage: {
          detail: '',
          operation: '',
          operationStatus: '',
          summary: '',
        },
      },
    });
  }
}
