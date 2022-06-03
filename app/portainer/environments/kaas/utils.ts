import { trackEvent } from '@/angulartics.matomo/analytics-services';
import { KaasProvider } from '@/portainer/settings/cloud/types';

import {
  CreateClusterFormValues,
  isAzureClusterFormValues,
  isGkeClusterFormValues,
  isEKSClusterFormValues,
} from './types';

export function sendKaasProvisionAnalytics(
  values: CreateClusterFormValues,
  provider: KaasProvider
) {
  trackEvent('portainer-endpoint-creation', {
    category: 'portainer',
    metadata: { type: 'agent', platform: 'kubernetes' },
  });

  if (isAzureClusterFormValues(values)) {
    trackEvent('provision-kaas-cluster', {
      category: 'kubernetes',
      metadata: {
        provider,
        region: values.region,
        'availability-zones': values.availabilityZones,
        teir: values.tier,
        'node-count': values.nodeCount,
        'node-size': values.nodeSize,
      },
    });
    return;
  }

  if (isGkeClusterFormValues(values)) {
    trackEvent('provision-kaas-cluster', {
      category: 'kubernetes',
      metadata: {
        provider,
        region: values.region,
        cpu: values.cpu,
        ram: values.ram,
        hdd: values.hdd,
        'node-size': values.nodeSize,
        'node-count': values.nodeCount,
      },
    });
    return;
  }

  if (isEKSClusterFormValues(values)) {
    trackEvent('provision-kaas-cluster', {
      category: 'kubernetes',
      metadata: {
        provider,
        region: values.region,
        'node-size': values.instanceType,
        'node-count': values.nodeCount,
      },
    });
    return;
  }

  // tracking for Linode, Civo and DigitalOcean
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider,
      region: values.region,
      'node-size': values.nodeSize,
      'node-count': values.nodeCount,
    },
  });
}
