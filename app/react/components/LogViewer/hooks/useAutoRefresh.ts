import { useEffect } from 'react';
import { UseQueryResult } from 'react-query';

import { AUTO_REFRESH_INTERVAL } from '@@/LogViewer/helpers/consts';

export function useAutoRefresh(
  autoRefresh: boolean,
  logsQuery: UseQueryResult
) {
  const { refetch } = logsQuery;

  useEffect(() => {
    let timer: NodeJS.Timeout;

    function startTimer() {
      timer = setTimeout(fetch, AUTO_REFRESH_INTERVAL);
    }

    async function fetch() {
      await refetch();
      startTimer();
    }

    if (autoRefresh) {
      startTimer();
    }

    return () => {
      if (timer) {
        clearTimeout(timer);
      }
    };
  }, [autoRefresh, refetch]);
}
