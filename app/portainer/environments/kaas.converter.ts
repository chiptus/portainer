import { Option } from '../components/form-components/Input/Select';

import { KaasInfo, KaasInfoResponse } from './components/kaas/kaas.types';

function buildOption(value: string, label?: string): Option<string> {
  return { value, label: label ?? value };
}

export function parseKaasInfoResponse(response: KaasInfoResponse): KaasInfo {
  const networks = response.Networks?.reduce((acc, region) => {
    acc.set(
      region.Region,
      region.Networks.map((net) => buildOption(net.Id, net.Name))
    );
    return acc;
  }, new Map<string, Option<string>[]>());

  return {
    networks,
    kubernetesVersions: response.KubernetesVersions,
    nodeSizes: response.NodeSizes.map((v) => buildOption(v.value, v.name)),
    regions: response.Regions.map((v) => buildOption(v.value, v.name)),
  };
}
