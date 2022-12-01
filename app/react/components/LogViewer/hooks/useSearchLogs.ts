import { useEffect, useState } from 'react';

import { LogInterface, ProcessedLogsInterface } from '@@/LogViewer/types';

function newSearchedLogs() {
  return {
    logs: [],
    totalKeywords: 0,
    keywordIndexes: [],
  } as ProcessedLogsInterface;
}

export function useSearchLogs(logs: LogInterface[], keyword: string) {
  const [searchedLogs, setSearchedLogs] = useState<ProcessedLogsInterface>(
    newSearchedLogs()
  );

  useEffect(() => {
    const searchedLogs: ProcessedLogsInterface = newSearchedLogs();

    if (keyword.length) {
      const lowerKeyword = keyword.toLowerCase();

      for (let i = 0; i < logs.length; i += 1) {
        const log = logs[i];

        const numOfKeywords = log.lowerLine.split(lowerKeyword).length - 1;
        searchedLogs.logs.push({
          ...log,
          numOfKeywords,
          firstKeywordIndex: searchedLogs.totalKeywords,
        });

        for (let j = 0; j < numOfKeywords; j += 1) {
          searchedLogs.keywordIndexes.push({
            lineNumber: i,
            indexInLine: j,
          });
        }

        searchedLogs.totalKeywords += numOfKeywords;
      }
    } else {
      searchedLogs.logs = logs;
    }

    setSearchedLogs(searchedLogs);
  }, [logs, keyword]);

  return { searchedLogs };
}
