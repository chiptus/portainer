import { object, number, string, SchemaOf } from 'yup';

import { CreateEKSClusterFormValues } from '../types';

export function validationSchema(): SchemaOf<CreateEKSClusterFormValues> {
  return object().shape({
    amiType: string().required('AMI type is required.'),
    instanceType: string().required('Instance type is required.'),
    nodeVolumeSize: number()
      .integer('Node volume size must be a whole number.')
      .min(1, 'Node volume size must be greater than or equal to 1 GiB.')
      .max(16384, 'Node volume size must be less than or equal to 16384 GiB.')
      .default(20),
  });
}
