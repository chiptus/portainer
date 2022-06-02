import { Option } from '@/portainer/components/form-components/Input/Select';

import {
  KaasInfo,
  KaasInfoResponse,
  isAzureKaasInfoResponse,
  NetworkInfo,
  isGKEKaasInfoResponse,
} from './types';

function buildOption(value: string, label?: string): Option<string> {
  return { value, label: label ?? value };
}

export function parseKaasInfoResponse(response: KaasInfoResponse): KaasInfo {
  const kubernetesVersions = response.kubernetesVersions.map((v) =>
    buildOption(v.value, v.name)
  );
  const regions = response.regions.map((v) => buildOption(v.value, v.name));

  // Azure response type
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

  const nodeSizes = response.nodeSizes.map((v) => buildOption(v.value, v.name));
  const networks =
    response.networks?.reduce((acc, n) => {
      const networkRegion = {
        [n.region]: n.networks.map((n) => buildOption(n.id, n.name)),
      } as NetworkInfo;
      return { ...acc, ...networkRegion };
    }, {} as NetworkInfo) || {};
  // GKE response type
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
