import { object, SchemaOf, string } from 'yup';

import { CreateKubeConfigEnvironment } from '@/react/portainer/environments/environment.service/create';

import { metadataValidation } from '../../shared/MetadataFieldset/validation';
import { useNameValidation } from '../../shared/NameField';

export function useValidation(): SchemaOf<CreateKubeConfigEnvironment> {
  return object({
    name: useNameValidation(),
    kubeConfig: string().required('Kubeconfig file is required.'),
    meta: metadataValidation(),
  });
}
