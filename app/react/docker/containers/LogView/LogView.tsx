import { useCallback } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { setPortainerAgentTargetHeader } from '@/portainer/services/http-request.helper';
import { getContainerLogs } from '@/react/docker/containers/containers.service';
import { useContainer } from '@/react/docker/containers/queries/container';
import { ContainerLogsParams } from '@/react/docker/containers/types';

import { PageHeader } from '@@/PageHeader';
import { LogViewer } from '@@/LogViewer';
import { GetLogsParamsInterface } from '@@/LogViewer/types';
import { InformationPanel } from '@@/InformationPanel';
import { TextTip } from '@@/Tip/TextTip';
import { Link } from '@@/Link';

export function LogView() {
  const {
    params: { endpointId: environmentId, id: containerId, nodeName },
  } = useCurrentStateAndParams();

  const getLogsFn = useCallback(
    (getLogsParams: GetLogsParamsInterface) => {
      setPortainerAgentTargetHeader(nodeName);

      const newParams: ContainerLogsParams = {
        stdout: true,
        stderr: true,
        timestamps: getLogsParams.timestamps,
        tail: getLogsParams.tail || 0,
        since: getLogsParams.since,
      };

      return getContainerLogs(environmentId, containerId, newParams);
    },
    [environmentId, containerId, nodeName]
  );

  const containerQuery = useContainer(environmentId, containerId);
  if (!containerQuery.data || containerQuery.isLoading) {
    return null;
  }
  const containerName = containerQuery.data.Name?.substring(1) || '';
  const logsEnabled =
    containerQuery.data.HostConfig?.LogConfig?.Type && // if a portion of the object path doesn't exist, logging is likely disabled
    containerQuery.data.HostConfig.LogConfig.Type !== 'none'; // if type === none logging is disabled
  const breadcrumbs = [
    { label: 'Containers', link: 'docker.containers' },
    {
      label: containerName,
      link: 'docker.containers.container',
      linkParams: { id: containerId },
    },
    'Logs',
  ];

  return (
    <>
      <PageHeader title="Container logs" breadcrumbs={breadcrumbs} reload />

      {logsEnabled ? (
        <LogViewer
          resourceType="docker-container"
          resourceName={containerName}
          getLogsFn={getLogsFn}
        />
      ) : (
        <LogsDisabledInfoPanel />
      )}
    </>
  );
}

function LogsDisabledInfoPanel() {
  const {
    params: { id: containerId, nodeName },
  } = useCurrentStateAndParams();

  return (
    <InformationPanel>
      <TextTip color="blue">
        Logging is disabled for this container. If you want to re-enable
        logging, please{' '}
        <Link
          to="docker.containers.new"
          params={{ from: containerId, nodeName }}
        >
          redeploy your container
        </Link>{' '}
        and select a logging driver in the &quot;Command & logging&quot; panel.
      </TextTip>
    </InformationPanel>
  );
}
