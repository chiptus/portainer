import { CreateNamespaceFormValues, CreateNamespacePayload } from './types';

export function transformFormValuesToNamespacePayload(
  createNamespaceFormValues: CreateNamespaceFormValues,
  owner: string
): CreateNamespacePayload {
  const memoryInBytes =
    Number(createNamespaceFormValues.resourceQuota.memory) * 10 ** 6;
  return {
    Name: createNamespaceFormValues.name,
    Owner: owner,
    Annotations: createNamespaceFormValues.annotations.reduce(
      (acc, { Key, Value }) => ({ ...acc, [Key]: Value }),
      {}
    ),
    ResourceQuota: {
      enabled: createNamespaceFormValues.resourceQuota.enabled,
      cpu: createNamespaceFormValues.resourceQuota.cpu,
      memory: `${memoryInBytes}`,
    },
    LoadBalancerQuota: createNamespaceFormValues.loadBalancerQuota,
    StorageQuotas: createNamespaceFormValues.storageQuota.reduce(
      (acc, storageClass) => ({
        ...acc,
        [storageClass.className]: {
          enabled: storageClass.enabled,
          limit: storageClass.size
            ? `${storageClass.size}${storageClass.sizeUnit}`
            : undefined,
        },
      }),
      {}
    ),
  };
}
