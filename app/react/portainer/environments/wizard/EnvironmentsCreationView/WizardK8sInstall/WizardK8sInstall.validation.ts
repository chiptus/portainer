import { number, object } from 'yup';

import { useNameValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';
import { metadataValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset/validation';

import { validationSchema as microk8sValidation } from './Microk8sCreateClusterForm/validation';
import { AddonOptionInfo } from './types';

export function useValidationSchema(addonOptions: AddonOptionInfo[]) {
  return object({
    name: useNameValidation()
      .matches(
        /^[a-z0-9-]+$/,
        'Name must only contain lowercase alphanumeric characters and hyphens.'
      )
      .max(32, 'Name must be 32 characters or less.')
      .matches(
        /^[a-z](?:[a-z0-9-]*[a-z0-9])?$/,
        'Name must start with a letter and end with an alphanumeric character.'
      ),
    meta: metadataValidation(),
    credentialId: number().required('Credentials are required.'),
    microk8s: microk8sValidation(addonOptions),
  });
}
