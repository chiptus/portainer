import { object, SchemaOf, string } from 'yup';

import { CreateKubeConfigEnvironment } from '@/portainer/environments/environment.service/create';

import { metadataValidation } from '../../shared/MetadataFieldset/validation';
import { nameValidation } from '../../shared/NameField';

export function validation(): SchemaOf<CreateKubeConfigEnvironment> {
  return object({
    name: nameValidation(),
    kubeConfig: string().required('Kubeconfig file is required.'),
    meta: metadataValidation(),
  });
}
