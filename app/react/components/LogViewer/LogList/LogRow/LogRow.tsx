import clsx from 'clsx';
import { useCallback, useEffect, useMemo, useRef } from 'react';

import { highlightKeyword } from '@@/LogViewer/helpers/highlightKeyword';
import { KeywordIndexInterface, LogInterface } from '@@/LogViewer/types';
import { LogSpan } from '@@/LogViewer/LogList/LogRow/LogSpan/LogSpan';
import { useLogViewerContext } from '@@/LogViewer/context';

import './LogRow.css';
import { useSetSizeContext } from '../useSetSize';

interface Props {
  lineIndex: number;
}

export function LogRow({ lineIndex }: Props) {
  const setSize = useSetSizeContext();
  const rowRef = useRef<HTMLDivElement | null>(null);
  const {
    logs,
    searchStatus: { focusedKeywordIndex },
    controllerStates: { wrapLine, keyword, showLineNumbers },
  } = useLogViewerContext();

  const log = logs.logs[lineIndex];
  const focusedKeywordIndexInLine = getFocusedKeywordIndexInLine(
    focusedKeywordIndex,
    logs.keywordIndexes,
    log
  );

  const handleSizeChange = useCallback(() => {
    setSize(lineIndex, rowRef.current?.getBoundingClientRect().height || 0);
  }, [lineIndex, setSize]);

  useEffect(() => {
    // Hack: wait for the row to be rendered (for rowRef to be set)
    setTimeout(() => handleSizeChange(), 250);
    window.addEventListener('resize', handleSizeChange);
    return () => window.removeEventListener('resize', handleSizeChange);
  }, [handleSizeChange, lineIndex, setSize, wrapLine]);

  const spans = useMemo(
    () => highlightKeyword(log, keyword, focusedKeywordIndexInLine),
    [log, keyword, focusedKeywordIndexInLine]
  );

  return (
    <pre className={clsx('log-row', { 'wrap-line': wrapLine })}>
      {showLineNumbers && (
        <div className="log-row-line-number">{log.lineNumber}</div>
      )}
      <div ref={rowRef} className={clsx('log-row-content')}>
        {spans.map(LogSpan)}
      </div>
    </pre>
  );
}

// Find out the keyword index in line of the focused keyword
// Return -1 if the focused keyword is not on current line
function getFocusedKeywordIndexInLine(
  focusedKeywordIndex: number,
  keywordIndexes: KeywordIndexInterface[],
  log: LogInterface
) {
  let focusedKeywordIndexInLine = -1;

  if (focusedKeywordIndex >= 0 && keywordIndexes[focusedKeywordIndex]) {
    if (keywordIndexes[focusedKeywordIndex].lineNumber + 1 === log.lineNumber) {
      focusedKeywordIndexInLine =
        keywordIndexes[focusedKeywordIndex].indexInLine;
    }
  }

  return focusedKeywordIndexInLine;
}
