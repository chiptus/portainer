import { object, number, string } from 'yup';

import { metadataValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset/validation';

export function validationSchema() {
  return object().shape({
    nodeCount: number()
      .integer('Node count must be a whole number.')
      .min(1, 'Node count must be greater than or equal to 1.')
      .max(450, 'Node count must be less than or equal to 450.')
      .required('Node count is required.'),
    amiType: string().required('AMI type is required.'),
    instanceType: string().required('Instance type is required.'),
    region: string().required('Region is required.'),
    kubernetesVersion: string(), // empty string is allowed for EKS\
    nodeVolumeSize: number()
      .integer('Node volume size must be a whole number.')
      .min(1, 'Node volume size must be greater than or equal to 1 GiB.')
      .max(16384, 'Node volume size must be less than or equal to 16384 GiB.'),
    meta: metadataValidation(),
  });
}
