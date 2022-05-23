import { object, string, number } from 'yup';

import { KaasProvider } from '@/portainer/settings/cloud/types';
import { metadataValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset/validation';
import { nameValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';

export function validationSchema() {
  return object().shape({
    name: nameValidation()
      .matches(
        /^[a-z0-9-]+$/,
        'Name must only contain lowercase alphanumeric characters and hyphens.'
      )
      .max(32, 'Name must be 32 characters or less.'),
    networkId: string().when('type', {
      is: KaasProvider.CIVO,
      then: string().required('Network ID is required.'),
    }),
    nodeCount: number()
      .min(1, 'Node count must be greater than or equal to 1.')
      .required('Node count is required.'),
    meta: metadataValidation(),
  });
}
