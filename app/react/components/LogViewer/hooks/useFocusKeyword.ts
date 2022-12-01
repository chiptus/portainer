import { VariableSizeList } from 'react-window';
import { RefObject, useContext, useEffect } from 'react';

import {
  LogViewerContext,
  LogViewerContextInterface,
} from '@@/LogViewer/types';

// scroll to focused keyword
export function useFocusKeyword(listRef: RefObject<VariableSizeList>) {
  const {
    logs,
    searchStatus: { focusedKeywordIndex },
  } = useContext(LogViewerContext) as LogViewerContextInterface;
  const { lineNumber } = logs.keywordIndexes[focusedKeywordIndex] || {};

  useEffect(() => {
    if (lineNumber !== undefined) {
      listRef.current?.scrollToItem(lineNumber);
    }
  }, [listRef, lineNumber]);
}
