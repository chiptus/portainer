import { useCallback } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import {
  getPodLogs,
  PodLogsParams,
} from '@/react/kubernetes/applications/axios/pods/getPodLogs';

import { PageHeader } from '@@/PageHeader';
import { LogViewer } from '@@/LogViewer';
import { GetLogsParamsInterface } from '@@/LogViewer/types';

export function LogView() {
  const {
    params: {
      endpointId: environmentId,
      container,
      name: appName,
      namespace,
      pod: podID,
    },
  } = useCurrentStateAndParams();

  const getLogsFn = useCallback(
    (getLogsParams: GetLogsParamsInterface) => {
      const newParams: PodLogsParams = {
        container,
        timestamps: getLogsParams.timestamps,
        tailLines: getLogsParams.tail || 0,
        sinceSeconds: getLogsParams.sinceSeconds,
      };

      return getPodLogs(environmentId, namespace, podID, newParams);
    },
    [environmentId, podID, container, namespace]
  );

  const breadcrumbs = [
    {
      label: 'Namespaces',
      link: 'kubernetes.resourcePools',
    },
    {
      label: namespace,
      link: 'kubernetes.resourcePools.resourcePool',
      linkParams: { id: namespace },
    },
    {
      label: 'Applications',
      link: 'kubernetes.applications',
    },
    {
      label: appName,
      link: 'kubernetes.applications.application',
      linkParams: { name: appName, namespace },
    },
    'Pods',
    podID,
    'Containers',
    container,
    'Logs',
  ];

  return (
    <>
      <PageHeader title="Application logs" breadcrumbs={breadcrumbs} />
      <LogViewer
        resourceType="kube-pod"
        resourceName={podID}
        getLogsFn={getLogsFn}
      />
    </>
  );
}
