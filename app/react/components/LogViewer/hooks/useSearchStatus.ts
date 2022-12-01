import { useEffect, useState } from 'react';

import { ProcessedLogsInterface } from '@@/LogViewer/types';
import { SearchStatusInterface } from '@@/LogViewer/LogController/SearchStatus/SearchStatus';

export function useSearchStatus(
  logs: ProcessedLogsInterface,
  visibleStartIndex: number
) {
  const [focusedKeywordIndex, setFocusedKeywordIndex] = useState<number>(-1);

  // Try to find out the nearest keyword after the line of #visibleStartIndex
  // And save its index at focusedKeywordIndex state
  useEffect(() => {
    let focusedKeywordIndex = -1;
    let focusLine = -1;

    for (let i = 0; i < logs.logs.length; i += 1) {
      const log = logs.logs[i];

      if (log.numOfKeywords) {
        if (focusLine < visibleStartIndex || focusedKeywordIndex === -1) {
          focusedKeywordIndex = log.firstKeywordIndex as number;
          focusLine = i;
        }
      }
    }
    setFocusedKeywordIndex(focusedKeywordIndex);

    // visibleStartIndex should not be a dependency here
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [logs.logs]);

  function nextKeyword() {
    setFocusedKeywordIndex((focusedKeywordIndex + 1) % logs.totalKeywords);
  }

  function previousKeyword() {
    setFocusedKeywordIndex(
      (focusedKeywordIndex - 1 + logs.totalKeywords) % logs.totalKeywords
    );
  }

  const searchStatus: SearchStatusInterface = {
    totalKeywords: logs.totalKeywords,
    focusedKeywordIndex,
    nextKeyword,
    previousKeyword,
  };

  return { searchStatus };
}
