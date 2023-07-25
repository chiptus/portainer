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
    name: string().notRequired().default(''),
    method: mixed().oneOf(Object.values(FormValuesFileMethod)).notRequired(),
    content: mixed().notRequired(),
  });
}
