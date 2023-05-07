import { K8sInstallFormValues, CreateMicrok8sClusterPayload } from './types';

export function formatMicrok8sPayload({
  microk8s: { nodeIPs, addons, kubernetesVersion },
  ...values
}: K8sInstallFormValues): CreateMicrok8sClusterPayload {
  const splitNodeIpsByCommas = nodeIPs.flatMap((ip) => ip.split(','));
  const cleanNodeIps = splitNodeIpsByCommas
    .map((ip) => ip.replaceAll(' ', ''))
    .filter((ip) => ip);
  return {
    ...values,
    nodeIPs: cleanNodeIps,
    addons,
    kubernetesVersion,
  };
}
