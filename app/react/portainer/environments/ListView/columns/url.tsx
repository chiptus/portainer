import { CellContext } from '@tanstack/react-table';
import { AlertCircle, HelpCircle, Loader2, Settings } from 'lucide-react';
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
      <>
        {environment.URL}
        {status !== '' && ( // status is in a provisioning or error state
          <div className="inline-block">
            <span
              className={clsx(
                'vertical-center inline-flex',
                status === 'error' ? 'text-warning' : 'text-muted'
              )}
            >
              {environment.URL && status === 'error' && (
                <Icon icon={AlertCircle} />
              )}
              {environment.URL && status === 'processing' && (
                <Icon icon={Loader2} className="animate-spin-slow" />
              )}

              <span>{environment.StatusMessage?.summary}</span>
            </span>
            <TooltipWithChildren
              message={
                <div>
                  <span
                    // eslint-disable-next-line react/no-danger
                    dangerouslySetInnerHTML={{
                      __html: sanitize(environment.StatusMessage?.detail ?? ''),
                    }}
                  />
                  {environment.URL && status === 'error' && (
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
              <span
                className={clsx(
                  'vertical-center inline-flex text-base',
                  status === 'error' ? 'text-warning' : 'text-muted'
                )}
              >
                <HelpCircle className="lucide ml-1" aria-hidden="true" />
              </span>
            </TooltipWithChildren>
          </div>
        )}
      </>
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
