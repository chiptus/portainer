import { VariableSizeList } from 'react-window';
import { RefObject, useEffect } from 'react';

import { useLogViewerContext } from '../context';

// scroll to focused keyword
export function useFocusKeyword(listRef: RefObject<VariableSizeList>) {
  const {
    logs,
    searchStatus: { focusedKeywordIndex },
  } = useLogViewerContext();
  const { lineNumber } = logs.keywordIndexes[focusedKeywordIndex] || {};

  useEffect(() => {
    if (lineNumber !== undefined) {
      listRef.current?.scrollToItem(lineNumber);
    }
  }, [listRef, lineNumber]);
}
