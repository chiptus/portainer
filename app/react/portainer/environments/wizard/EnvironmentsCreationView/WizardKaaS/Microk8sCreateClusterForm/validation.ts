import { object, string, number, array, SchemaOf } from 'yup';

import { CreateMicrok8sClusterFormValues } from '../types';

export function validationSchema(): SchemaOf<CreateMicrok8sClusterFormValues> {
  return object().shape({
    nodeIP1: string().required('Node IP is required.'),
    nodeIP2: string().default(''),
    nodeIP3: string().default(''),
    customTemplateId: number().default(0),
    addons: array(),
  });
}
