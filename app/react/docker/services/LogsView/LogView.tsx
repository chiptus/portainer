import { useCallback } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { setPortainerAgentTargetHeader } from '@/portainer/services/http-request.helper';
import { useService } from '@/react/docker/services/queries/service';
import { getServiceLogs } from '@/react/docker/services/axios/getServiceLogs';
import { ServiceLogsParams } from '@/react/docker/services/types';

import { PageHeader } from '@@/PageHeader';
import { LogViewer } from '@@/LogViewer';
import { GetLogsParamsInterface } from '@@/LogViewer/types';

export function LogView() {
  const {
    params: { endpointId: environmentId, id: serviceId, nodeName },
  } = useCurrentStateAndParams();

  const getLogsFn = useCallback(
    (getLogsParams: GetLogsParamsInterface) => {
      setPortainerAgentTargetHeader(nodeName);

      const newParams: ServiceLogsParams = {
        stdout: true,
        stderr: true,
        timestamps: getLogsParams.timestamps,
        tail: getLogsParams.tail || 0,
        since: getLogsParams.since,
      };

      return getServiceLogs(environmentId, serviceId, newParams);
    },
    [environmentId, serviceId, nodeName]
  );

  const serviceQuery = useService(environmentId, serviceId);
  const serviceName = serviceQuery.data?.Spec.Name || '';

  const breadcrumbs = [
    { label: 'Services', link: 'docker.services' },
    {
      label: serviceName,
      link: 'docker.services.service',
      linkParams: { id: serviceId },
    },
    'Logs',
  ];

  return (
    <>
      <PageHeader title="Service logs" breadcrumbs={breadcrumbs} reload />
      <LogViewer
        resourceType="docker-service"
        resourceName={serviceName}
        getLogsFn={getLogsFn}
      />
    </>
  );
}
