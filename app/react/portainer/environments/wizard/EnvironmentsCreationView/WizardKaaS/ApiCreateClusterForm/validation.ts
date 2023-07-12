import { object, string, SchemaOf } from 'yup';

import { KaasProvider } from '@/react/portainer/environments/types';

import { CreateApiClusterFormValues } from '../types';

export function validationSchema(): SchemaOf<CreateApiClusterFormValues> {
  return object().shape({
    networkId: string()
      .when('type', {
        is: KaasProvider.CIVO,
        then: string().required('Network ID is required.'),
      })
      .default(''),
    nodeSize: string().required('Node size is required.'),
  });
}
