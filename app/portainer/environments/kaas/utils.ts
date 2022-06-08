import { trackEvent } from '@/angulartics.matomo/analytics-services';
import { KaasProvider } from '@/portainer/settings/cloud/types';

import { FormValues } from './types';

export function sendKaasProvisionAnalytics(
  values: FormValues,
  provider: KaasProvider
) {
  trackEvent('portainer-endpoint-creation', {
    category: 'portainer',
    metadata: { type: 'agent', platform: 'kubernetes' },
  });

  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      trackGoogleCloudProvision(values);
      break;
    case KaasProvider.CIVO:
    case KaasProvider.LINODE:
    case KaasProvider.DIGITAL_OCEAN:
      trackApiProvision(provider, values);
      break;
    case KaasProvider.AZURE:
      trackAzureProvision(values);
      break;
    case KaasProvider.AWS:
      trackAmazonProvision(values);
      break;
    default:
      break;
  }
}

function trackAzureProvision(values: FormValues) {
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider: KaasProvider.AZURE,
      region: values.region,
      'availability-zones': values.azure.availabilityZones,
      teir: values.azure.tier,
      'node-count': values.nodeCount,
      'node-size': values.nodeSize,
    },
  });
}

function trackGoogleCloudProvision(values: FormValues) {
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider: KaasProvider.GOOGLE_CLOUD,
      region: values.region,
      cpu: values.google.cpu,
      ram: values.google.ram,
      hdd: values.google.hdd,
      'node-size': values.nodeSize,
      'node-count': values.nodeCount,
    },
  });
}

function trackApiProvision(provider: KaasProvider, values: FormValues) {
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

function trackAmazonProvision(values: FormValues) {
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider: KaasProvider.AWS,
      region: values.region,
      'node-size': values.amazon.instanceType,
      'node-count': values.nodeCount,
    },
  });
}
