import { boolean, number, object, SchemaOf } from 'yup';

import { FormValues } from './types';

export function validationSchema(): SchemaOf<FormValues> {
  return object().shape({
    EnableEdgeComputeFeatures: boolean().required('This field is required.'),
    EnforceEdgeID: boolean().required('This field is required.'),
    EdgeAgentCheckinInterval: number().default(0),
  });
}
