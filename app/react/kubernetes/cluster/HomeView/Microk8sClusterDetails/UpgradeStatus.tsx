import { useEffect, useState } from 'react';
import { Loader2, X } from 'lucide-react';
import clsx from 'clsx';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { useUpdateEnvironmentMutation } from '@/react/portainer/environments/queries/useUpdateEnvironmentMutation';

import { AlertContainer, alertSettings } from '@@/Alert/Alert';
import { Icon } from '@@/Icon';
import { TextTip } from '@@/Tip/TextTip';
import { Tooltip } from '@@/Tip/Tooltip';
import { Button } from '@@/buttons';

export function UpgradeStatus() {
  const [autoRefreshRate, setAutoRefreshRate] = useState<number | undefined>();
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId, (env) => env, {
    autoRefreshRate,
  });
  const operationStatus = environment?.StatusMessage?.operationStatus;
  const updateEnvironmentStatusMutation = useUpdateEnvironmentMutation();

  // if operationStatus is processing, change the autorefresh rate to 10000ms, otherwise set it to undefined
  useEffect(() => {
    if (operationStatus === 'processing') {
      setAutoRefreshRate(10000);
      return;
    }
    setAutoRefreshRate(undefined);
  }, [operationStatus]);

  // '' is an idle status
  if (operationStatus === '') {
    return null;
  }

  return (
    <AlertContainer
      className={
        operationStatus === 'error'
          ? alertSettings.error.container
          : alertSettings.info.container
      }
    >
      <div className="flex flex-col gap-y-4 text-sm">
        <div className={clsx('flex items-center justify-between')}>
          <div
            className={clsx(
              'flex items-center',
              operationStatus === 'error'
                ? alertSettings.error.body
                : alertSettings.info.body
            )}
          >
            {operationStatus === 'processing' && (
              <Icon icon={Loader2} className="!mr-2 animate-spin-slow" />
            )}
            {environment?.StatusMessage?.summary}
            <Tooltip
              message={environment?.StatusMessage?.detail}
              position="top"
            />
          </div>
          {operationStatus === 'error' && (
            <Button
              icon={X}
              className={clsx(
                'flex !text-gray-7 hover:!text-gray-8 th-highcontrast:!text-gray-6 th-highcontrast:hover:!text-gray-5 th-dark:!text-gray-6 th-dark:hover:!text-gray-5'
              )}
              color="link"
              size="small"
              onClick={() => dismissMessage()}
            />
          )}
        </div>
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
