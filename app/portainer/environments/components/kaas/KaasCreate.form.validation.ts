import { object, string, number } from 'yup';

import { KaasProvider } from '@/portainer/settings/cloud/types';

export function validationSchema(environmentNames: string[]) {
  return object().shape({
    name: string()
      .required('Name is required.')
      .matches(
        /^[a-z0-9-]+$/,
        'Name must only contain lowercase alphanumeric characters and hyphens.'
      )
      .test('not in array', 'Name is already in use.', (newEnvironmentName) =>
        environmentNames.every((name) => name !== newEnvironmentName)
      )
      .max(32, 'Name must be 32 characters or less.'),
    networkId: string().when('type', {
      is: KaasProvider.CIVO,
      then: string().required('Network ID is required.'),
    }),
    nodeCount: number()
      .min(1, 'Node count must be greater than or equal to 1.')
      .required('Node count is required.'),
  });
}
