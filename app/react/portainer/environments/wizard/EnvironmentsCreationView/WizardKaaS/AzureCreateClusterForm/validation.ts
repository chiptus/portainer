import { object, string, array, SchemaOf } from 'yup';

import { CreateAzureClusterFormValues } from '../types';

export function validationSchema(): SchemaOf<CreateAzureClusterFormValues> {
  return object().shape({
    resourceGroup: string()
      .default('')
      .when('resourceGroupInput', {
        is: 'select',
        then: string().required(
          'No resource groups available in the selected region, please change region or add a resource group.'
        ),
      }),
    resourceGroupName: string()
      .default('')
      .when('resourceGroupInput', {
        is: 'input',
        then: string()
          .required('Resource group name is required.')
          .matches(
            /^[a-z0-9-_]+$/,
            'Resource group name must only contain lowercase alphanumeric characters and hyphens.'
          )
          .max(90, 'Resource group name must be 90 characters or less.'),
      }),
    poolName: string()
      .required('Pool name is required.')
      .matches(
        /^[a-z0-9]+$/,
        'Pool name must only contain lowercase alphanumeric characters.'
      )
      .max(11, 'Pool name must be 11 characters or less.'),
    nodeSize: string().required('Node size is required.'),
    availabilityZones: array().of(string()).default([]),
    tier: string().required('Tier is required.'),
    dnsPrefix: string()
      .required('DNS prefix is required.')
      .max(54, 'DNS prefix must be 54 characters or less.')
      .matches(
        /^[a-zA-Z0-9-]+$/,
        'DNS prefix name must only contain alphanumeric characters and hyphens.'
      )
      .matches(
        /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/,
        'DNS prefix name must start and end with an alphanumeric character.'
      ),
    resourceGroupInput: string().oneOf(['select', 'input']).default('select'),
  });
}
