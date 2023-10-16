import { useCurrentStateAndParams } from '@uirouter/react';

import { EdgeConfigurationCategory } from '@/react/edge/edge-configurations/types';

export function useCategory() {
  const {
    params: { category = EdgeConfigurationCategory.Configuration },
  } = useCurrentStateAndParams();

  return [category];
}
