import { K8sInstallFormValues, CreateMicrok8sClusterPayload } from './types';

export function formatMicrok8sPayload({
  microk8s: { nodeIPs, addons, customTemplateId, kubernetesVersion },
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
    customTemplateId,
    kubernetesVersion,
  };
}
