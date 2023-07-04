import { object, number, string, array, SchemaOf, TestContext } from 'yup';

import { AddNodesFormValues } from '@/react/kubernetes/cluster/NodeCreateView/types';

import { CreateMicrok8sClusterFormValues } from '../types';

export function validationSchema(): SchemaOf<CreateMicrok8sClusterFormValues> {
  return object().shape({
    masterNodes: validateNodeIPList().test(
      'first line not empty',
      'At least one control plane node is required',
      validateFirstLine
    ),
    workerNodes: validateNodeIPList(), // worker nodes can be empty on creation
    customTemplateId: number().default(0),
    addons: array(),
    kubernetesVersion: string().required('version is required'),
  });
}

export function validateNodeIPList(existingNodeIPAddresses?: string[]) {
  return array()
    .test(
      'valid IPV4',
      'Must have a valid IP address or address range separated by commas.',
      validateIpList
    )
    .test(
      'no duplicate IPs in input',
      'Duplicate IPs are not allowed',
      validateNoDuplicatesInInput
    )
    .test(
      'no duplicate IPs in form',
      'Duplicate IPs are not allowed',
      validateNoDuplicateIPsInForm
    )
    .test(
      'no duplicates with existing nodes',
      "An IP address you're trying to add is already assigned to an existing node in the cluster. Please use a different IP address.",
      (value) => validateNoDuplicateIPsInCluster(value, existingNodeIPAddresses)
    );
}

function validateFirstLine(ipList?: string[]) {
  return !!ipList && ipList.length > 0 && !!ipList[0];
}

function validateIpList(ipList?: string[]) {
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

function validateNoDuplicatesInInput(ipList?: string[]) {
  if (!ipList || ipList.length === 0) {
    return true;
  }

  const ipListSeparatedByIp = formatSeparateIPAddresses(ipList);

  // if the length of the set is the same as the length of the array, there are no duplicates
  const ipSet = new Set(ipListSeparatedByIp);
  return ipSet.size === ipListSeparatedByIp.length;
}

/** validateNoDuplicateIPsInForm checks for duplicates between both masterNodesToAdd and workerNodesToAdd input */
function validateNoDuplicateIPsInForm(this: TestContext) {
  const { masterNodesToAdd, workerNodesToAdd } = this
    .parent as Partial<AddNodesFormValues>;
  const ipList = [...(masterNodesToAdd ?? []), ...(workerNodesToAdd ?? [])];
  if (ipList.length === 0) {
    return true;
  }

  const ipListSeparatedByIp = formatSeparateIPAddresses(ipList);

  // if the length of the set is the same as the length of the array, there are no duplicates
  const ipSet = new Set(ipListSeparatedByIp);
  return ipSet.size === ipListSeparatedByIp.length;
}

function validateNoDuplicateIPsInCluster(
  formIPList?: string[],
  nodeIPList?: string[]
) {
  if (
    !formIPList ||
    formIPList?.length === 0 ||
    !nodeIPList ||
    nodeIPList?.length === 0
  ) {
    return true;
  }

  const ipListSeparatedByIp = formatSeparateIPAddresses(formIPList);
  const nodeIPListContainsAnyFormIP = ipListSeparatedByIp.some((ip) =>
    nodeIPList.includes(ip)
  );
  return !nodeIPListContainsAnyFormIP;
}

/** formatSeparateIPAddresses takes a list of IP addresses and IP address ranges from the form, which can be on the same line, and returns a list of all the separated IP addresses in the form */
function formatSeparateIPAddresses(formIPList: string[]) {
  return formIPList
    .flatMap((ip) => ip?.split(/,|-/)) // split by comma or dash
    .map((ip) => ip?.trim()) // trim whitespace
    .map((ip) => ip?.replace(/(^|\.)0+(\d)/g, '$1$2')) // remove all leading 0's from each octet
    .filter((ip) => ip); // remove any empty or undefined lines
}
