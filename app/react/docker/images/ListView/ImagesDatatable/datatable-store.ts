import create from 'zustand';
import { persist } from 'zustand/middleware';

import { keyBuilder } from '@/react/hooks/useLocalStorage';
import {
  paginationSettings,
  sortableSettings,
  refreshableSettings,
} from '@/react/components/datatables/types';

import { TableSettings } from './types';

export function createStore(storageKey: string) {
  return create<TableSettings>()(
    persist(
      (set) => ({
        ...sortableSettings(set),
        ...paginationSettings(set),
        ...refreshableSettings(set),
      }),
      {
        name: keyBuilder(storageKey),
      }
    )
  );
}
