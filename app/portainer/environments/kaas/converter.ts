import { Option } from '@/portainer/components/form-components/Input/Select';

import { KaasInfo, KaasInfoResponse, isGKEKaasInfoResponse } from './types';

function buildOption(value: string, label?: string): Option<string> {
  return { value, label: label ?? value };
}

export function parseKaasInfoResponse(response: KaasInfoResponse): KaasInfo {
  const networks =
    response.networks?.reduce((acc, region) => {
      acc.set(
        region.region,
        region.networks.map((net) => buildOption(net.id, net.name))
      );
      return acc;
    }, new Map<string, Option<string>[]>()) ||
    new Map<string, Option<string>[]>();

  const kubernetesVersions = response.kubernetesVersions.map((v) =>
    buildOption(v.value, v.name)
  );
  const nodeSizes = response.nodeSizes.map((v) => buildOption(v.value, v.name));
  const regions = response.regions.map((v) => buildOption(v.value, v.name));

  // if GKE response type
  if (isGKEKaasInfoResponse(response)) {
    return {
      ...response,
      kubernetesVersions,
      nodeSizes,
      networks,
      regions,
    };
  }
  return {
    networks,
    nodeSizes,
    regions,
    kubernetesVersions,
  };
}
