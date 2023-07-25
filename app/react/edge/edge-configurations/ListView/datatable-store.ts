import {
  createPersistedStore,
  BasicTableSettings,
} from '@/react/components/datatables/types';

type TableSettings = BasicTableSettings;

export function createStore(storageKey: string) {
  return createPersistedStore<TableSettings>(storageKey);
}
