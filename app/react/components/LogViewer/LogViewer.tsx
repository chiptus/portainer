import { useMemo, useRef, useState } from 'react';

import { LogList } from '@@/LogViewer/LogList/LogList';
import { LogController } from '@@/LogViewer/LogController/LogController';
import {
  GetLogsFnType,
  LogViewerContext,
  LogViewerContextInterface,
} from '@@/LogViewer/types';
import { TableContainer } from '@@/datatables';
import { useSearchStatus } from '@@/LogViewer/hooks/useSearchStatus';
import { useSearchLogs } from '@@/LogViewer/hooks/useSearchLogs';
import { useFilterLogs } from '@@/LogViewer/hooks/useFilterLogs';
import { useAutoRefresh } from '@@/LogViewer/hooks/useAutoRefresh';
import { useFetchLogs } from '@@/LogViewer/hooks/usFetchLogs';
import { useLogsQuery } from '@@/LogViewer/hooks/useLogsQuery';
import { useControllerStates } from '@@/LogViewer/hooks/useControllerStates';

interface Props {
  getLogsFn: GetLogsFnType;
  hideFetch?: boolean;
  hideLines?: boolean;
  hideTimestamp?: boolean;
  resourceName: string;
  resourceType: string;
}

export function LogViewer({
  getLogsFn,
  hideFetch,
  hideLines,
  hideTimestamp,
  resourceName,
  resourceType,
}: Props) {
  const logViewerRef = useRef<HTMLDivElement>(null);
  const [visibleStartIndex, setVisibleStartIndex] = useState(0);

  const { controllerStates } = useControllerStates();
  const { logsQuery, originalLogs } = useLogsQuery(
    getLogsFn,
    controllerStates,
    resourceType,
    resourceName
  );

  useAutoRefresh(controllerStates.autoRefresh, logsQuery);

  const { searchedLogs } = useSearchLogs(
    originalLogs,
    controllerStates.keyword
  );

  const { filteredLogs } = useFilterLogs(
    searchedLogs,
    controllerStates.filter,
    controllerStates.keyword
  );

  const { searchStatus } = useSearchStatus(filteredLogs, visibleStartIndex);

  useFetchLogs(controllerStates, logsQuery);

  const context = useMemo<LogViewerContextInterface>(
    () => ({
      controllerStates,
      logViewerRef,
      logs: filteredLogs,
      visibleStartIndex,
      setVisibleStartIndex,
      searchStatus,
      hideFetch,
      hideLines,
      hideTimestamp,
      resourceName,
      logsQuery,
    }),
    [
      controllerStates,
      logViewerRef,
      filteredLogs,
      visibleStartIndex,
      searchStatus,
      hideFetch,
      hideLines,
      hideTimestamp,
      resourceName,
      logsQuery,
    ]
  );

  return (
    <LogViewerContext.Provider value={context}>
      <div className="col-sm-12" ref={logViewerRef}>
        <TableContainer>
          <LogController />
          <LogList logs={filteredLogs} />
        </TableContainer>
      </div>
    </LogViewerContext.Provider>
  );
}
