import { SchemaOf, array, object, string, boolean, mixed } from 'yup';

import { StorageQuotaFormValues } from './types';

export const storageQuotaValidationSchema: SchemaOf<StorageQuotaFormValues[]> =
  array(
    object({
      className: string().required('Storage quota name is required.'),
      enabled: boolean().required('Storage quota enabled status is required.'),
      size: string().when('enabled', {
        is: true,
        then: string()
          .test(
            'is-valid-size',
            'Storage quota must be a positive number.',
            (value) => {
              if (!value) {
                return true;
              }
              const parsedValue = Number(value);
              return !Number.isNaN(parsedValue) && parsedValue >= 0;
            }
          )
          .required('Size is required.'),
      }),
      sizeUnit: mixed()
        .oneOf(['M', 'G', 'T'])
        .required('Storage quota unit is required.'),
    })
  );
