import { boolean, object, SchemaOf, number } from 'yup';

import { LoadBalancerQuotaFormValues } from './types';

export const loadBalancerValidationSchema: SchemaOf<LoadBalancerQuotaFormValues> =
  object({
    enabled: boolean().required('Load balancer enabled status is required.'),
    limit: number().when('enabled', {
      is: true,
      then: number()
        .min(0, 'Max load balancers value must be a positive number.')
        .required('Max load balancers value is required.'),
    }),
  });
