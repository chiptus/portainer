import { useMemo, useState } from 'react';

import { useLocalStorage } from './useLocalStorage';

export function usePaginationLimitState(
  key: string
): [number, (value: number) => void] {
  const paginationKey = paginationKeyBuilder(key);
  const [pageLimit, setPageLimit] = useLocalStorage(paginationKey, 10);

  return [pageLimit, setPageLimit];

  function paginationKeyBuilder(key: string) {
    return `datatable_pagination_${key}`;
  }
}

export function usePaginationState(key: string): {
  page: number;
  setPage: (value: number) => void;
  pageLimit: number;
  setPageLimit: (value: number) => void;
} {
  const [page, setPage] = useState(0);
  const [pageLimit, setPageLimit] = usePaginationLimitState(key);

  const state = useMemo(
    () => ({
      page,
      setPage,
      pageLimit,
      setPageLimit,
    }),
    [page, setPage, pageLimit, setPageLimit]
  );

  return state;
}
