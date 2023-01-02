import { VariableSizeList } from 'react-window';
import { RefObject, useCallback, useEffect, useState } from 'react';

import { useLogViewerContext } from '../context';

export function useKeepAtBottom(
  outerRef: RefObject<HTMLDivElement>,
  listRef: RefObject<VariableSizeList>
) {
  const [isScrollAtBottom, setIsScrollAtBottom] = useState(false);
  const [keepAtBottom, setKeepAtBottom] = useState(false);

  const { logs, setVisibleStartIndex } = useLogViewerContext();

  const checkIsScrollAtBottom = useCallback(() => {
    const ref = outerRef.current;
    if (ref) {
      const isBottom = ref.scrollHeight - ref.scrollTop - ref.clientHeight < 10;
      setIsScrollAtBottom(isBottom);
      return isBottom;
    }
    return false;
  }, [outerRef]);

  const onItemsRendered = useCallback(
    ({ visibleStartIndex }) => {
      setVisibleStartIndex(visibleStartIndex);
      checkIsScrollAtBottom();
    },
    [checkIsScrollAtBottom, setVisibleStartIndex]
  );

  const onScroll = useCallback(
    ({ scrollUpdateWasRequested }) => {
      const isAtBottom = checkIsScrollAtBottom();
      if (!scrollUpdateWasRequested) {
        setKeepAtBottom(isAtBottom);
      }
    },
    [checkIsScrollAtBottom]
  );

  const logLength = logs.logs.length;

  const scrollToBottom = useCallback(() => {
    setKeepAtBottom(true);
    listRef.current?.scrollToItem(logLength);
  }, [logLength, listRef]);

  useEffect(() => {
    if (keepAtBottom) {
      listRef.current?.scrollToItem(logLength);
    }
  }, [keepAtBottom, listRef, logLength]);

  return { isScrollAtBottom, onItemsRendered, onScroll, scrollToBottom };
}
