import { useState } from 'react';
import moment from 'moment';
import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';

import { formatLogs } from '@@/LogViewer/helpers/formatLogs';
import {
  ControllerStatesInterface,
  GetLogsFnType,
  GetLogsParamsInterface,
  LogInterface,
} from '@@/LogViewer/types';

const AUTO_REFRESH_INTERVAL = 5000;

export function useLogsQuery(
  getLogsFn: GetLogsFnType,
  { tail, showTimestamp, since, autoRefresh }: ControllerStatesInterface,
  resourceType: string,
  resourceName: string
) {
  const [originalLogs, setOriginalLogs] = useState<LogInterface[]>([]);

  function queryFn() {
    const getLogsParams: GetLogsParamsInterface = {
      tail,
      timestamps: showTimestamp,
      since: since ? moment().subtract(since, 'seconds').unix() : 0,
      sinceSeconds: since || 0,
    };

    return getLogsFn(getLogsParams);
  }

  const logsQuery = useQuery(['logs', resourceType, resourceName], queryFn, {
    onSuccess: (logs) => {
      const formattedLogs = formatLogs(logs);
      setOriginalLogs(formattedLogs);
    },
    refetchInterval: () => (autoRefresh ? AUTO_REFRESH_INTERVAL : false),
    refetchOnWindowFocus: false,
    ...withError('Failure', 'Unable to get logs.'),
  });

  return { logsQuery, originalLogs };
}
