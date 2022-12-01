import { useCallback, useRef } from 'react';

import { SizeMapInterface } from '@@/LogViewer/types';
import { DEFAULT_ROW_SIZE } from '@@/LogViewer/helpers/consts';

export function useSizeCache(listReset: () => void) {
  const sizeCache = useRef<SizeMapInterface>({});

  const setSize = useCallback(
    (index, size) => {
      sizeCache.current = { ...sizeCache.current, [index]: size };
      listReset();
    },
    [listReset]
  );

  const getSize = useCallback(
    (index: number) => sizeCache.current[index] || DEFAULT_ROW_SIZE,
    []
  );

  return { setSize, getSize };
}
