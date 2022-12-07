import createStore from 'zustand';
import { persist } from 'zustand/middleware';

import { keyBuilder } from '@/react/hooks/useLocalStorage';
import { TagId } from '@/portainer/tags/types';
import { EnvironmentGroupId } from '@/react/portainer/environments/environment-groups/types';
import {
  PlatformType,
  EnvironmentStatus,
} from '@/react/portainer/environments/types';

import { ConnectionType } from './types';

export interface Filters {
  platformTypes: Array<PlatformType>;
  connectionTypes: Array<ConnectionType>;
  status: Array<EnvironmentStatus>;
  tagIds?: Array<TagId>;
  groupIds: Array<EnvironmentGroupId>;
  agentVersions: Array<string>;
  sort?: string;
  sortDesc: boolean;
}

export const useFiltersStore = createStore<{
  value: Filters;
  handleChange(value: Partial<Filters>): void;
  clear(): void;
}>()(
  persist(
    (set) => ({
      value: getDefaultValue(),
      handleChange: (value: Partial<Filters>) => {
        set((state) => ({
          value: {
            ...state.value,
            ...value,
          },
        }));
      },
      clear: () => {
        set({
          value: getDefaultValue(),
        });
      },
    }),
    {
      name: keyBuilder('datatable_home_filter'),
      getStorage: () => localStorage,
    }
  )
);

function getDefaultValue() {
  return {
    agentVersions: [],
    connectionTypes: [],
    groupIds: [],
    platformTypes: [],
    sortDesc: false,
    sort: undefined,
    status: [],
    tagIds: undefined,
  };
}
