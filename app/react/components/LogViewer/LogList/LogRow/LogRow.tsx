import clsx from 'clsx';
import { useContext, useEffect, useMemo, useRef } from 'react';

import { highlightKeyword } from '@@/LogViewer/helpers/highlightKeyword';
import {
  KeywordIndexInterface,
  LogInterface,
  LogViewerContext,
  LogViewerContextInterface,
} from '@@/LogViewer/types';
import { LogSpan } from '@@/LogViewer/LogList/LogRow/LogSpan/LogSpan';

import './LogRow.css';

interface Props {
  lineIndex: number;
  setSize: (index: number, size: number) => void;
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

export function LogRow({ setSize, lineIndex }: Props) {
  const rowRef = useRef<HTMLDivElement | null>(null);
  const {
    logs,
    searchStatus: { focusedKeywordIndex },
    controllerStates: { wrapLine, keyword, showLineNumbers },
  } = useContext(LogViewerContext) as LogViewerContextInterface;

  const log = logs.logs[lineIndex];
  const focusedKeywordIndexInLine = getFocusedKeywordIndexInLine(
    focusedKeywordIndex,
    logs.keywordIndexes,
    log
  );

  useEffect(() => {
    setSize(lineIndex, rowRef.current?.getBoundingClientRect().height || 0);
  }, [lineIndex, wrapLine, setSize]);

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
