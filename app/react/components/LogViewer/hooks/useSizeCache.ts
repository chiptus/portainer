import { useCallback, useRef } from 'react';

interface SizeMapInterface {
  [key: number]: number;
}

const DEFAULT_ROW_SIZE = 25;

export function useSizeCache(listReset: () => void) {
  const sizeCache = useRef<SizeMapInterface>({});

  const setSize = useCallback(
    (index: number, size: number) => {
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
