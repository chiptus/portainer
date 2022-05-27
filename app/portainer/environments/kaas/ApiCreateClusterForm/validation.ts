import { object, string, number } from 'yup';

import { metadataValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset/validation';
import { KaasProvider } from '@/portainer/settings/cloud/types';

export function validationSchema() {
  return object().shape({
    networkId: string().when('type', {
      is: KaasProvider.CIVO,
      then: string().required('Network ID is required.'),
    }),
    nodeCount: number()
      .integer('Node count must be an integer.')
      .min(1, 'Node count must be greater than or equal to 1.')
      .max(1000, 'Node count must be less than or equal to 1000.')
      .required('Node count is required.'),
    meta: metadataValidation(),
  });
}
