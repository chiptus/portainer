import _ from 'lodash';
import { useCallback } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { PodLogsParams } from '@/react/kubernetes/applications/axios/pods/getPodLogs';

import { PageHeader } from '@@/PageHeader';
import { LogViewer } from '@@/LogViewer';
import { GetLogsFnType, GetLogsParamsInterface } from '@@/LogViewer/types';

interface Props {
  getLogsFn: GetLogsFnType;
}

export function LogView({ getLogsFn }: Props) {
  const {
    params: { name, namespace },
  } = useCurrentStateAndParams();

  const getLogsWithParamsFn = useCallback(
    (getLogsParams: GetLogsParamsInterface) => {
      const newParams: PodLogsParams = {
        timestamps: getLogsParams.timestamps,
        tailLines: getLogsParams.tail || 0,
        sinceSeconds: getLogsParams.sinceSeconds,
      };

      return getLogsFn(_.pickBy(newParams));
    },
    [getLogsFn]
  );

  const breadcrumbs = [
    { label: 'Namespaces', link: 'kubernetes.resourcePools' },
    {
      label: namespace,
      link: 'kubernetes.resourcePools.resourcePool',
      linkParams: { id: namespace },
    },
    { label: 'Applications', link: 'kubernetes.applications' },
    'Stacks',
    name,
    'Logs',
  ];

  return (
    <div>
      <PageHeader title="Stack logs" breadcrumbs={breadcrumbs} />
      <LogViewer
        resourceType="kube-stack"
        resourceName={name}
        getLogsFn={getLogsWithParamsFn}
      />
    </div>
  );
}
