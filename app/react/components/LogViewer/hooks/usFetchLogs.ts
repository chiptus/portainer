import _ from 'lodash';
import { UseQueryResult } from 'react-query';
import { useEffect, useMemo } from 'react';

import { ControllerStatesInterface } from '@@/LogViewer/types';

export function useFetchLogs(
  { tail, showTimestamp, since }: ControllerStatesInterface,
  logsQuery: UseQueryResult
) {
  const debounceRefetch = useMemo(
    () => _.debounce(logsQuery.refetch, 300),
    [logsQuery.refetch]
  );

  useEffect(() => {
    debounceRefetch();
  }, [tail, since, showTimestamp, debounceRefetch]);
}
