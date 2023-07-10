import { EnvironmentId } from '../types';

export const environmentQueryKeys = {
  base: () => ['environments'] as const,
  item: (id: EnvironmentId) => [...environmentQueryKeys.base(), id] as const,
};
