import { HelpCircle, AlertCircle, Settings } from 'lucide-react';

import { notifySuccess } from '@/portainer/services/notifications';

import { Button } from '@@/buttons';
import { TooltipWithChildren } from '@@/Tip/TooltipWithChildren';

import { Environment, EnvironmentStatusMessage } from '../types';
import { useUpdateEndpointStatusMessageMutation } from '../queries/useUpdateEnvironmentMutation';

export interface Props {
  endpoint: Environment;
  onReload?: () => void;
  className?: string;
}

export function EnvironmentURL({ endpoint, onReload }: Props) {
  const mutation = useUpdateEndpointStatusMessageMutation();

  if (endpoint.Type !== 4 && endpoint.Status !== 3) {
    return (
      <>
        {endpoint.URL}
        {endpoint.StatusMessage.Summary && endpoint.StatusMessage.Detail && (
          <div className="ml-2 inline-block">
            <span className="text-danger vertical-center inline-flex">
              <AlertCircle className="lucide" aria-hidden="true" />
              <span>{endpoint.StatusMessage.Summary}</span>
            </span>
            <TooltipWithChildren
              message={
                <div>
                  {endpoint.StatusMessage.Detail}
                  {endpoint.URL && (
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
              <span className="vertical-center inline-flex text-base">
                <HelpCircle className="lucide ml-1" aria-hidden="true" />
              </span>
            </TooltipWithChildren>
          </div>
        )}
      </>
    );
  }

  if (endpoint.Type === 4) {
    return <>-</>;
  }

  if (endpoint.Status === 3) {
    const status = (
      <span className="vertical-center inline-flex text-base">
        <Settings className="lucide animate-spin-slow" />
        {endpoint.StatusMessage.Summary}
      </span>
    );
    if (!endpoint.StatusMessage.Detail) {
      return status;
    }
    return (
      <TooltipWithChildren
        message={endpoint.StatusMessage.Detail}
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
        environmentId: endpoint.Id,
        payload: {
          IsSetStatusMessage: true,
          StatusMessage: {} as EnvironmentStatusMessage,
        },
      },
      {
        onSuccess: () => {
          notifySuccess('Success', 'Error dismissed successfully');
          if (onReload) {
            onReload();
          }
        },
      }
    );
  }
}
