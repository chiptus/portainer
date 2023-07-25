import { object, SchemaOf, string, mixed } from 'yup';

import { FormValues, FormValuesFileMethod } from '../common/types';
import { validationBase } from '../common/validation';

export function validation(): SchemaOf<FormValues> {
  return object({
    file: fileValidation(),
  }).concat(validationBase());
}

function fileValidation(): SchemaOf<FormValues['file']> {
  return object({
    name: string().required('This field is required'),
    method: mixed().oneOf(Object.values(FormValuesFileMethod)).defined(),
    content: mixed().required('Content can not be empty'),
  });
}
