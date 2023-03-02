import { boolean, number, object, string, SchemaOf } from 'yup';

import { options as asyncIntervalOptions } from '@/react/edge/components/EdgeAsyncIntervalsForm';

import { file, withFileSize } from '@@/form-components/yup-file-validation';

import { FormValues } from './types';

const intervals = asyncIntervalOptions.map((option) => option.value);
const MAX_FILE_SIZE = 5_242_880; // 5MB

export function validationSchema(): SchemaOf<FormValues> {
  return object({
    EdgeAgentCheckinInterval: number().required('This field is required.'),
    Edge: object({
      PingInterval: number()
        .required('This field is required.')
        .oneOf(intervals),
      SnapshotInterval: number()
        .required('This field is required.')
        .oneOf(intervals),
      CommandInterval: number()
        .required('This field is required.')
        .oneOf(intervals),
      AsyncMode: boolean().default(false),
      MTLS: object({
        UseSeparateCert: boolean().default(false),
        CaCertFile: certValidation(),
        CertFile: certValidation(),
        KeyFile: certValidation(),
        CaCert: string().default(''),
        Cert: string().default(''),
        Key: string().default(''),
      }),
      TunnelServerAddress: string().required('This field is required.'),
    }),
  });
}

function certValidation() {
  return withFileSize(file(), MAX_FILE_SIZE).when(['UseSeparateCert'], {
    is: (UseSeparateCert: boolean) => UseSeparateCert,
    then: (schema) => schema.required('File is required'),
  });
}
