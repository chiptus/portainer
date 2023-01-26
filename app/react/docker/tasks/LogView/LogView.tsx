import { useCallback } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { getTaskLogs } from '@/react/docker/tasks/axios/getTaskLogs';
import { setPortainerAgentTargetHeader } from '@/portainer/services/http-request.helper';
import { useTask } from '@/react/docker/tasks/queries/task';
import { useService } from '@/react/docker/services/queries/service';
import { TaskLogsParams } from '@/react/docker/tasks/types';

import { PageHeader } from '@@/PageHeader';
import { LogViewer } from '@@/LogViewer';
import { GetLogsParamsInterface } from '@@/LogViewer/types';

export function LogView() {
  const {
    params: { endpointId: environmentId, id: taskId, nodeName },
  } = useCurrentStateAndParams();

  const getLogsFn = useCallback(
    (getLogsParams: GetLogsParamsInterface) => {
      setPortainerAgentTargetHeader(nodeName);

      const newParams: TaskLogsParams = {
        stdout: true,
        stderr: true,
        timestamps: getLogsParams.timestamps,
        tail: getLogsParams.tail || 0,
        since: getLogsParams.since,
      };

      return getTaskLogs(environmentId, taskId, newParams);
    },
    [environmentId, taskId, nodeName]
  );

  const taskQuery = useTask(environmentId, taskId);
  const taskID = taskQuery.data?.ID || '';
  const serviceID = taskQuery.data?.ServiceID || '';

  const serviceQuery = useService(environmentId, serviceID);
  const serviceName = serviceQuery.data?.Spec.Name || '';

  const breadcrumbs = [
    { label: 'Services', link: 'docker.services' },
    {
      label: serviceName,
      link: 'docker.services.service',
      linkParams: { id: serviceID },
    },
    { label: taskID, link: 'docker.tasks.task', linkParams: taskId },
    'Logs',
  ];

  return (
    <>
      <PageHeader title="Task logs" breadcrumbs={breadcrumbs} />
      <LogViewer
        resourceType="docker-task"
        resourceName={taskID}
        getLogsFn={getLogsFn}
      />
    </>
  );
}
