import { SchemaOf, array, object, string, boolean, number, mixed } from 'yup';

import { StorageQuotaFormValues } from './types';

export const storageQuotaValidationSchema: SchemaOf<StorageQuotaFormValues[]> =
  array(
    object({
      className: string().required('Storage quota name is required.'),
      enabled: boolean().required('Storage quota enabled status is required.'),
      size: number().when('enabled', {
        is: true,
        then: number()
          .min(0, 'Storage quota must be a positive number.')
          .required('Storage quota is required.'),
      }),
      sizeUnit: mixed()
        .oneOf(['M', 'G', 'T'])
        .required('Storage quota unit is required.'),
    })
  );
