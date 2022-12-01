import { useEffect, useState } from 'react';

import { ProcessedLogsInterface } from '@@/LogViewer/types';

export function useFilterLogs(
  logs: ProcessedLogsInterface,
  filter: boolean,
  keyword: string
) {
  const [filteredLogs, setFilteredLogs] = useState<ProcessedLogsInterface>({
    ...logs,
  });

  useEffect(() => {
    const newFilteredLogs: ProcessedLogsInterface = { ...logs };

    if (keyword && filter) {
      newFilteredLogs.logs = logs.logs.filter((v) => v.numOfKeywords);
    }

    setFilteredLogs(newFilteredLogs);
  }, [logs, filter, keyword]);

  return { filteredLogs };
}
