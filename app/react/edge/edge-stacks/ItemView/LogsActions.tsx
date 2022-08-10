import clsx from 'clsx';

import { notifySuccess } from '@/portainer/services/notifications';
import { EnvironmentId } from '@/portainer/environments/types';

import { Button } from '@@/buttons';

import { useCollectLogsMutation } from '../queries/useCollectLogsMutation';
import { useDeleteLogsMutation } from '../queries/useDeleteLogsMutation';
import { useDownloadLogsMutation } from '../queries/useDownloadLogsMutation';
import { EdgeStack } from '../types';
import { useLogsStatus } from '../queries/useLogsStatus';

interface Props {
  environmentId: EnvironmentId;
  edgeStackId: EdgeStack['Id'];
}

export function LogsActions({ environmentId, edgeStackId }: Props) {
  const logsStatusQuery = useLogsStatus(edgeStackId, environmentId);
  const collectLogsMutation = useCollectLogsMutation();
  const downloadLogsMutation = useDownloadLogsMutation();
  const deleteLogsMutation = useDeleteLogsMutation();

  if (!logsStatusQuery.isSuccess) {
    return null;
  }

  const status = logsStatusQuery.data;

  const collecting = collectLogsMutation.isLoading || status === 'pending';

  return (
    <>
      <Button color="none" title="Retrieve logs" onClick={handleCollectLogs}>
        <i
          className={clsx('fa', {
            'fa-file-alt': !collecting,
            'fa-circle-notch fa-spin': collecting,
          })}
          aria-hidden="true"
        />
      </Button>
      <Button
        color="none"
        title="Download logs"
        disabled={status !== 'collected'}
        onClick={handleDownloadLogs}
      >
        <i
          className={clsx('fa', {
            'fa-cloud-download-alt': !downloadLogsMutation.isLoading,
            'fa-circle-notch fa-spin': downloadLogsMutation.isLoading,
          })}
          aria-hidden="true"
        />
      </Button>
      <Button
        color="none"
        title="Delete logs"
        disabled={status !== 'collected'}
        onClick={handleDeleteLogs}
      >
        <i
          className={clsx('fa', {
            'fa-backspace': !deleteLogsMutation.isLoading,
            'fa-circle-notch fa-spin': deleteLogsMutation.isLoading,
          })}
          aria-hidden="true"
        />
      </Button>
    </>
  );

  function handleCollectLogs() {
    if (status === 'pending') {
      return;
    }

    collectLogsMutation.mutate(
      {
        edgeStackId,
        environmentId,
      },
      {
        onSuccess() {
          notifySuccess('Success', 'Logs Collection started');
        },
      }
    );
  }

  function handleDownloadLogs() {
    downloadLogsMutation.mutate({
      edgeStackId,
      environmentId,
    });
  }

  function handleDeleteLogs() {
    deleteLogsMutation.mutate(
      {
        edgeStackId,
        environmentId,
      },
      {
        onSuccess() {
          notifySuccess('Success', 'Logs Deleted');
        },
      }
    );
  }
}
