import { KaasProvider } from '@/portainer/settings/cloud/types';

import { Option } from '@@/form-components/Input/Select';

import {
  KaasInfo,
  KaasInfoResponse,
  isAzureKaasInfoResponse,
  NetworkInfo,
  isGKEKaasInfoResponse,
  isEksKaasInfoResponse,
  CreateApiClusterPayload,
  CreateAzureClusterPayload,
  CreateEksClusterPayload,
  CreateGKEClusterPayload,
  FormValues,
} from './types';

function buildOption(value: string, label?: string): Option<string> {
  return { value, label: label ?? value };
}

export function parseKaasInfoResponse(response: KaasInfoResponse): KaasInfo {
  const kubernetesVersions = response.kubernetesVersions.map((v) =>
    buildOption(v.value, v.name)
  );
  const regions = response.regions.map((v) => buildOption(v.value, v.name));

  if (isAzureKaasInfoResponse(response)) {
    return {
      ...response,
      kubernetesVersions,
      resourceGroups:
        response.resourceGroups?.map((rg) => buildOption(rg, rg)) || [],
      regions,
      nodeSizes: response.nodeSizes,
    };
  }

  if (isEksKaasInfoResponse(response)) {
    return {
      kubernetesVersions,
      instanceTypes: response.instanceTypes,
      regions,
      amiTypes: response.amiTypes.map((v) => buildOption(v.value, v.name)),
    };
  }

  const nodeSizes = response.nodeSizes.map((v) => buildOption(v.value, v.name));
  const networks =
    response.networks?.reduce((acc, n) => {
      const networkRegion = {
        [n.region]: n.networks.map((n) => buildOption(n.id, n.name)),
      } as NetworkInfo;
      return { ...acc, ...networkRegion };
    }, {} as NetworkInfo) || {};

  if (isGKEKaasInfoResponse(response)) {
    return {
      ...response,
      kubernetesVersions,
      nodeSizes,
      networks,
      regions,
    };
  }

  // API response type
  return {
    networks,
    nodeSizes,
    regions,
    kubernetesVersions,
  };
}

export function getPayloadParse(provider: KaasProvider) {
  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      return googlePayload;
    case KaasProvider.CIVO:
    case KaasProvider.LINODE:
    case KaasProvider.DIGITAL_OCEAN:
      return apiPayload;
    case KaasProvider.AZURE:
      return azurePayload;
    case KaasProvider.AWS:
      return amazonPayload;
    default:
      throw new Error('Unsupported provider');
  }

  function googlePayload({
    azure,
    api,
    amazon,
    google: { cpu, ram, ...google },
    ...values
  }: FormValues): CreateGKEClusterPayload {
    if (google.nodeSize === 'custom') {
      return { cpu, ram, ...google, ...values };
    }

    return { ...google, ...values };
  }

  function apiPayload({
    amazon,
    azure,
    google,
    api,
    ...values
  }: FormValues): CreateApiClusterPayload {
    return { ...api, ...values };
  }

  function azurePayload({
    amazon,
    azure,
    google,
    api,
    ...values
  }: FormValues): CreateAzureClusterPayload {
    let resourceGroup = '';
    let { resourceGroupName } = azure;
    if (azure.resourceGroupInput === 'select') {
      resourceGroup = azure.resourceGroup;
      resourceGroupName = '';
    }

    return {
      ...azure,
      ...values,
      resourceGroup,
      resourceGroupName,
    };
  }

  function amazonPayload({
    amazon,
    azure,
    google,
    api,
    ...values
  }: FormValues): CreateEksClusterPayload {
    return {
      ...amazon,
      ...values,
    };
  }
}
