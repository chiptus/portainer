export function formatNodeIPs(nodeIPs: string[]) {
  return nodeIPs
    .flatMap((ip) => ip.split(','))
    .map((ip) => ip.replaceAll(' ', ''))
    .filter((ip) => ip);
}
