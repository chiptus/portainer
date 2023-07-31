import { CellContext } from '@tanstack/react-table';
import {
  AlertCircle,
  AlertTriangle,
  HelpCircle,
  Loader2,
  Settings,
} from 'lucide-react';
import clsx from 'clsx';
import sanitize from 'sanitize-html';

import { EnvironmentStatus } from '@/react/portainer/environments/types';
import { notifySuccess } from '@/portainer/services/notifications';
import { PortainerEndpointTypes } from '@/portainer/models/endpoint/models';

import { TooltipWithChildren } from '@@/Tip/TooltipWithChildren';
import { Button } from '@@/buttons';
import { Icon } from '@@/Icon';

import { EnvironmentListItem } from '../types';
import { useUpdateEnvironmentMutation } from '../../queries/useUpdateEnvironmentMutation';

import { columnHelper } from './helper';

export const url = columnHelper.accessor('URL', {
  header: 'URL',
  cell: Cell,
});

function Cell({
  row: { original: environment },
}: CellContext<EnvironmentListItem, string>) {
  const mutation = useUpdateEnvironmentMutation();
  const status = environment.StatusMessage?.operationStatus;

  if (
    environment.Type !== PortainerEndpointTypes.EdgeAgentOnDockerEnvironment &&
    environment.Status !== EnvironmentStatus.Provisioning
  ) {
    return (
      <div className="inline-flex gap-2 whitespace-nowrap">
        {environment.URL && <div>{environment.URL}</div>}
        {status !== '' && ( // status is in a provisioning or error state
          <div
            className={clsx('vertical-center flex items-center', {
              'text-danger': status === 'error',
              'text-warning': status === 'warning',
              'text-muted': status !== 'error' && status !== 'warning',
            })}
          >
            {environment.URL && status === 'error' && (
              <Icon className="flex-none" icon={AlertCircle} />
            )}
            {environment.URL && status === 'warning' && (
              <Icon className="flex-none" icon={AlertTriangle} />
            )}
            {environment.URL && status === 'processing' && (
              <Icon className="flex-none animate-spin-slow" icon={Loader2} />
            )}
            {environment.StatusMessage?.summary}
            <TooltipWithChildren
              message={
                <div>
                  <span
                    // eslint-disable-next-line react/no-danger
                    dangerouslySetInnerHTML={{
                      __html: sanitize(environment.StatusMessage?.detail ?? ''),
                    }}
                  />
                  {environment.URL && status === 'warning' && (
                    <div className="mt-2 text-right">
                      <Button
                        color="link"
                        className="small !ml-0 p-0"
                        onClick={handleDismissButton}
                      >
                        <span className="text-muted-light">
                          Dismiss error (still visible in logs)
                        </span>
                      </Button>
                    </div>
                  )}
                </div>
              }
              position="bottom"
            >
              <span className="vertical-center text-muted inline-flex whitespace-nowrap text-base">
                <HelpCircle className="lucide" aria-hidden="true" />
              </span>
            </TooltipWithChildren>
          </div>
        )}
      </div>
    );
  }

  if (
    environment.Type === PortainerEndpointTypes.EdgeAgentOnDockerEnvironment
  ) {
    return <>-</>;
  }

  if (environment.Status === EnvironmentStatus.Provisioning) {
    const status = (
      <span className="vertical-center inline-flex text-base">
        <Settings className="lucide animate-spin-slow" />
        {environment.StatusMessage?.summary}
      </span>
    );
    if (!environment.StatusMessage?.detail) {
      return status;
    }
    return (
      <TooltipWithChildren
        message={environment.StatusMessage?.detail}
        position="bottom"
      >
        {status}
      </TooltipWithChildren>
    );
  }

  return <>-</>;

  function handleDismissButton() {
    mutation.mutate(
      {
        id: environment.Id,
        payload: {
          IsSetStatusMessage: true,
          StatusMessage: {
            detail: '',
            operation: '',
            operationStatus: '',
            summary: '',
          },
        },
      },
      {
        onSuccess: () => {
          notifySuccess('Success', 'Error dismissed successfully');
        },
      }
    );
  }
}
