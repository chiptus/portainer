import { array, object, SchemaOf, string, number, mixed } from 'yup';

import {
  FormValues,
  FormValuesEdgeConfigurationMatchingRule,
  FormValuesEdgeConfigurationType,
} from './types';

export function validationBase(): SchemaOf<Omit<FormValues, 'file'>> {
  return object({
    groupIds: array()
      .of(number().default(0))
      .min(1, 'At least one group is required'),
    name: string().required('This field is required'),
    directory: string().required('This field is required'),
    type: mixed<FormValuesEdgeConfigurationType>()
      .oneOf(Object.values(FormValuesEdgeConfigurationType))
      .required(),
    matchingRule: mixed<FormValuesEdgeConfigurationMatchingRule>()
      .oneOf(Object.values(FormValuesEdgeConfigurationMatchingRule))
      .when('type', {
        is: FormValuesEdgeConfigurationType.DeviceSpecific,
        then: (schema) => schema.required('This field is required'),
        otherwise: (schema) => schema.notRequired(),
      }),
  });
}
