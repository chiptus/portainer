import create from 'zustand';
import { persist } from 'zustand/middleware';

import { ContainersTableSettings } from '@/react/docker/containers/types';
import { keyBuilder } from '@/portainer/hooks/useLocalStorage';

export interface TableSettings extends ContainersTableSettings {
  setHiddenColumns: (hiddenColumns: string[]) => void;
  setSortBy: (id: string, desc: boolean) => void;
  setPageSize: (size: number) => void;
  truncateContainerName: number;
  setTruncateContainerName: (value: number) => void;
}

export function createStore(storageKey: string) {
  return create<TableSettings>()(
    persist(
      (set) => ({
        sortBy: { id: 'name', desc: false },
        pageSize: 10,
        setSortBy: (id: string, desc: boolean) => set({ sortBy: { id, desc } }),
        setPageSize: (pageSize: number) => set({ pageSize }),
        setTruncateContainerName: (truncateContainerName: number) =>
          set({
            truncateContainerName,
          }),
        truncateContainerName: 32,
        hiddenColumns: [],
        setHiddenColumns: (hiddenColumns: string[]) => set({ hiddenColumns }),
        autoRefreshRate: 0,
        hiddenQuickActions: [],
      }),
      {
        name: keyBuilder(storageKey),
      }
    )
  );
}
