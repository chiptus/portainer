import { object, number, array, mixed, SchemaOf } from 'yup';

import { CreateMicrok8sClusterFormValues, Microk8sK8sVersion } from '../types';

export function validationSchema(): SchemaOf<CreateMicrok8sClusterFormValues> {
  return object().shape({
    nodeIPs: array()
      .test(
        'valid IPV4',
        'Must have a valid IP address or address range separated by commas.',
        validateIpList
      )
      .test('first line not empty', 'Node IP is required', validateFirstLine)
      .test(
        'no duplicate IPs',
        'Duplicate IPs are not allowed',
        validateNoDuplicateIPs
      ),
    customTemplateId: number().default(0),
    addons: array(),
    kubernetesVersion: mixed<Microk8sK8sVersion>().required(
      'Kubernetes version is required'
    ),
  });
}

// I don't want to use any[] as the parameter type, but it is required by yup
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function validateFirstLine(ipList?: any[]) {
  return !!ipList && ipList.length > 0 && !!ipList[0];
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function validateIpList(ipList?: any[]) {
  // remove any empty or undefined lines
  const ipListNoEmptyLines = ipList?.filter((ip) => ip);

  // validate each ip address or range to match the regex (ipv4 only)
  const requiredIPV4Regex =
    /^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})([\s,-]+\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})*$/;
  return (
    !!ipListNoEmptyLines &&
    ipListNoEmptyLines.every((ip) => requiredIPV4Regex.test(ip.trim()))
  );
}

function validateNoDuplicateIPs(ipList?: string[]) {
  if (!ipList || ipList.length === 0) {
    return true;
  }

  const ipListSeparatedByIp = ipList
    .flatMap((ip) => ip?.split(/,|-/)) // split by comma or dash
    .map((ip) => ip?.trim()) // trim whitespace
    .map((ip) => ip?.replace(/(^|\.)0+(\d)/g, '$1$2')) // remove all leading 0's from each octet
    .filter((ip) => ip); // remove any empty or undefined lines

  // if the length of the set is the same as the length of the array, there are no duplicates
  const ipSet = new Set(ipListSeparatedByIp);
  return ipSet.size === ipListSeparatedByIp.length;
}
