import { useCallback } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { PageHeader } from '@@/PageHeader';
import { LogViewer } from '@@/LogViewer';
import { RawLogsSection } from '@@/LogViewer/types';

import { getTaskLogs } from './logs.service';

export function LogsView() {
  const {
    params: {
      endpointId: environmentId,
      jobID,
      taskName,
      allocationID,
      namespace,
    },
  } = useCurrentStateAndParams();

  const getLogsFn = useCallback(async () => {
    const stderrPromise = getTaskLogs(
      environmentId,
      namespace,
      jobID,
      allocationID,
      taskName,
      'stderr'
    );
    const stdoutPromise = getTaskLogs(
      environmentId,
      namespace,
      jobID,
      allocationID,
      taskName,
      'stdout'
    );
    const [stderrData, stdoutData] = await Promise.all([
      stderrPromise,
      stdoutPromise,
    ]);

    const ret: RawLogsSection[] = [];
    if (stderrData)
      ret.push({
        sectionName: 'STDERR',
        sectionNameColor: 'red',
        logs: stderrData,
      });
    if (stdoutData)
      ret.push({
        sectionName: 'STDOUT',
        sectionNameColor: 'orange',
        logs: stdoutData,
      });

    return ret;
  }, [environmentId, namespace, jobID, allocationID, taskName]);

  const breadcrumbs = [
    { label: 'Nomad Jobs', link: 'nomad.jobs' },
    jobID,
    taskName,
    'Logs',
  ];

  return (
    <>
      <PageHeader title="Task logs" breadcrumbs={breadcrumbs} />
      <LogViewer
        resourceType="nomad-task"
        resourceName={taskName}
        getLogsFn={getLogsFn}
        hideFetch
        hideLines
        hideTimestamp
      />
    </>
  );
}
