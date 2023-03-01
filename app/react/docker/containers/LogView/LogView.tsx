import { useCallback } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { setPortainerAgentTargetHeader } from '@/portainer/services/http-request.helper';
import { getContainerLogs } from '@/react/docker/containers/containers.service';
import { useContainer } from '@/react/docker/containers/queries/container';
import { ContainerLogsParams } from '@/react/docker/containers/types';

import { PageHeader } from '@@/PageHeader';
import { LogViewer } from '@@/LogViewer';
import { GetLogsParamsInterface } from '@@/LogViewer/types';

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
  const containerName = containerQuery.data?.Name.substring(1) || '';
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
      <PageHeader title="Container logs" breadcrumbs={breadcrumbs} />
      <LogViewer
        resourceType="docker-container"
        resourceName={containerName}
        getLogsFn={getLogsFn}
      />
    </>
  );
}
