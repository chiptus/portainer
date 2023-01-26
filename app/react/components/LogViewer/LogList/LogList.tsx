import { useResizeDetector } from 'react-resize-detector';
import { VariableSizeList } from 'react-window';
import AutoSizer from 'react-virtualized-auto-sizer';
import { CSSProperties, useCallback, useEffect, useRef } from 'react';

import { BottomButton } from '@@/LogViewer/LogList/BottomButton';
import { ProcessedLogsInterface } from '@@/LogViewer/types';
import { LogRow } from '@@/LogViewer/LogList/LogRow/LogRow';
import { useSizeCache } from '@@/LogViewer/hooks/useSizeCache';
import { useFocusKeyword } from '@@/LogViewer/hooks/useFocusKeyword';
import { useKeepAtBottom } from '@@/LogViewer/hooks/useKeepAtBottom';

import { SetSizeProvider } from './useSetSize';

interface Props {
  logs: ProcessedLogsInterface;
}

export function LogList({ logs }: Props) {
  const listRef = useRef<VariableSizeList>(null);
  const outerRef = useRef<HTMLDivElement>(null);

  const listReset = useCallback(() => listRef.current?.resetAfterIndex(0), []);

  const { getSize, setSize } = useSizeCache(listReset);

  useFocusKeyword(listRef);

  const { isScrollAtBottom, onItemsRendered, onScroll, scrollToBottom } =
    useKeepAtBottom(outerRef, listRef);

  const { width } = useResizeDetector({ targetRef: outerRef });
  useEffect(() => {
    listReset();
  }, [listReset, width]);

  return (
    <>
      <AutoSizer defaultHeight={750} disableWidth className="h-full">
        {({ height }) => (
          <SetSizeProvider setSize={setSize}>
            <VariableSizeList
              height={height}
              itemCount={logs.logs.length}
              itemSize={getSize}
              width="100%"
              onScroll={onScroll}
              onItemsRendered={onItemsRendered}
              ref={listRef}
              outerRef={outerRef}
              className="log-list"
            >
              {renderItem}
            </VariableSizeList>
          </SetSizeProvider>
        )}
      </AutoSizer>
      <BottomButton visible={!isScrollAtBottom} onClick={scrollToBottom} />
    </>
  );
}

function renderItem({ index, style }: { index: number; style: CSSProperties }) {
  return (
    <div style={{ ...style, width: 'unset' }}>
      <LogRow lineIndex={index} />
    </div>
  );
}
