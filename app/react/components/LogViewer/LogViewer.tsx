import { useMemo, useRef, useState } from 'react';

import { LogList } from '@@/LogViewer/LogList/LogList';
import { LogController } from '@@/LogViewer/LogController/LogController';
import { GetLogsFnType, LogViewerContextInterface } from '@@/LogViewer/types';
import { Widget } from '@@/Widget';

import { useSearchStatus } from './hooks/useSearchStatus';
import { useSearchLogs } from './hooks/useSearchLogs';
import { useFilterLogs } from './hooks/useFilterLogs';
import { useFetchLogs } from './hooks/usFetchLogs';
import { useLogsQuery } from './hooks/useLogsQuery';
import { useControllerStates } from './hooks/useControllerStates';
import { LogViewerProvider } from './context';

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
    <LogViewerProvider value={context}>
      <div className="col-sm-12" ref={logViewerRef}>
        <Widget className="h-full">
          <LogController />
          <Widget.Body className="no-padding">
            <LogList logs={filteredLogs} />
          </Widget.Body>
        </Widget>
      </div>
    </LogViewerProvider>
  );
}
