import { useCurrentStateAndParams } from '@uirouter/react';

import { EdgeConfigurationCategoryString } from '@/react/edge/edge-configurations/queries/create/types';

export function useCategory() {
  const {
    params: { category = EdgeConfigurationCategoryString.Configuration },
  } = useCurrentStateAndParams();

  return [category];
}
