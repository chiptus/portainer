import { mixed, number, object, SchemaOf, string } from 'yup';

import { KaasProvider } from '@/portainer/settings/cloud/types';
import { nameValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';
import { metadataValidation } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset/validation';

import { validationSchema as gkeValidation } from './GKECreateClusterForm/validation';
import { validationSchema as apiValidation } from './ApiCreateClusterForm/validation';
import { validationSchema as azureValidation } from './AzureCreateClusterForm/validation';
import { validationSchema as amazonValidation } from './EKSCreateClusterForm/validation';
import { FormValues, KaasInfo } from './types';

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
        /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/,
        'Name must start and end with an alphanumeric character.'
      ),
    meta: metadataValidation(),
    api: providerValidation(
      provider,
      [KaasProvider.CIVO, KaasProvider.DIGITAL_OCEAN, KaasProvider.LINODE],
      apiValidation()
    ),
    azure: providerValidation(
      provider,
      [KaasProvider.AZURE],
      azureValidation()
    ),
    google: providerValidation(
      provider,
      [KaasProvider.GOOGLE_CLOUD],
      gkeValidation(kaasInfo)
    ),
    kubernetesVersion: string()
      .default('')
      .when({
        is: provider !== KaasProvider.AWS,
        then: string().required('Kubernetes version is required.'),
      }),
    credentialId: number().required('Credentials are required.'),
    nodeCount: number()
      .integer('Node count must be an integer.')
      .min(1, 'Node count must be greater than or equal to 1.')
      .max(1000, 'Node count must be less than or equal to 1000.')
      .required('Node count is required.'),
    nodeSize: string().default(''),
    region: string().default(''),
    amazon: providerValidation(
      provider,
      [KaasProvider.AWS],
      amazonValidation()
    ),
  });
}

function providerValidation<T>(
  provider: KaasProvider,
  oneOf: KaasProvider[],
  validation: SchemaOf<T>
) {
  return mixed().when({
    is: oneOf.includes(provider),
    then: validation,
  });
}
