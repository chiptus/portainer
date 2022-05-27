import { trackEvent } from '@/angulartics.matomo/analytics-services';
import { KaasProvider } from '@/portainer/settings/cloud/types';

import {
  CreateApiClusterFormValues,
  CreateGKEClusterFormValues,
} from './types';

export function sendKaasProvisionAnalytics(
  values: CreateApiClusterFormValues | CreateGKEClusterFormValues,
  provider: KaasProvider
) {
  trackEvent('portainer-endpoint-creation', {
    category: 'portainer',
    metadata: { type: 'agent', platform: 'kubernetes' },
  });
  if ('cpu' in values) {
    // tracking for GKE
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
  } else {
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
}
