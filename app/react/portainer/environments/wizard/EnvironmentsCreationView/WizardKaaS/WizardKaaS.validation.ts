import { number, object, string, mixed, SchemaOf } from 'yup';

import { KaasProvider } from '@/portainer/settings/cloud/types';
import { nameValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';
import { metadataValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset/validation';

import { validationSchema as gkeValidation } from './GKECreateClusterForm/validation';
import { validationSchema as apiValidation } from './ApiCreateClusterForm/validation';
import { validationSchema as azureValidation } from './AzureCreateClusterForm/validation';
import { validationSchema as amazonValidation } from './EKSCreateClusterForm/validation';
import { KaasInfo, KaaSFormType, FormValues } from './types';
import { providerFormType } from './utils';

export function validationSchema(
  provider: KaasProvider,
  kaasInfo?: KaasInfo | null
): SchemaOf<Omit<FormValues, 'api' | 'azure' | 'google' | 'amazon'>> {
  return object({
    name: nameValidation()
      .matches(
        /^[a-z0-9-]+$/,
        'Name must only contain lowercase alphanumeric characters and hyphens.'
      )
      .max(32, 'Name must be 32 characters or less.')
      .matches(
        /^[a-z](?:[a-z0-9-]*[a-z0-9])?$/,
        'Name must start with a letter and end with an alphanumeric character.'
      ),
    meta: metadataValidation(),
    api:
      providerFormType(provider) === KaaSFormType.API
        ? apiValidation()
        : mixed(),
    azure:
      providerFormType(provider) === KaaSFormType.AZURE
        ? azureValidation()
        : mixed(),
    google:
      providerFormType(provider) === KaaSFormType.GKE
        ? gkeValidation(kaasInfo)
        : mixed(),
    kubernetesVersion: string()
      .default('')
      .when({
        is: provider !== KaasProvider.AWS,
        then: string().required('Kubernetes version is required.'),
      }),
    credentialId: number().required('Credentials are required.'),
    region: string().default(''),
    nodeCount: number()
      .integer('Node count must be a whole number.')
      .min(1, 'Node count must be greater than or equal to 1.')
      .max(1000, 'Node count must be less than or equal to 1000.')
      .required('Node count is required.'),
    amazon:
      providerFormType(provider) === KaaSFormType.EKS
        ? amazonValidation()
        : mixed(),
  });
}
