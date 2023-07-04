import { TrackEventProps } from '@/angulartics.matomo/analytics-services';

import { KaasProvider } from '../WizardK8sInstall/types';

import { FormValues, KaaSFormType } from './types';

export function sendKaasProvisionAnalytics(
  values: FormValues,
  provider: KaasProvider,
  trackEvent: (action: string, properties: TrackEventProps) => void
) {
  trackEvent('portainer-endpoint-creation', {
    category: 'portainer',
    metadata: { type: 'agent', platform: 'kubernetes' },
  });

  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      trackGoogleCloudProvision(values, trackEvent);
      break;
    case KaasProvider.CIVO:
    case KaasProvider.LINODE:
    case KaasProvider.DIGITAL_OCEAN:
      trackApiProvision(provider, values, trackEvent);
      break;
    case KaasProvider.AZURE:
      trackAzureProvision(values, trackEvent);
      break;
    case KaasProvider.AWS:
      trackAmazonProvision(values, trackEvent);
      break;
    default:
      break;
  }
}

function trackAzureProvision(
  values: FormValues,
  trackEvent: (action: string, properties: TrackEventProps) => void
) {
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider: KaasProvider.AZURE,
      region: values.region,
      'availability-zones': values.azure.availabilityZones,
      teir: values.azure.tier,
      'node-count': values.nodeCount,
      'node-size': values.azure.nodeSize,
    },
  });
}

function trackGoogleCloudProvision(
  values: FormValues,
  trackEvent: (action: string, properties: TrackEventProps) => void
) {
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider: KaasProvider.GOOGLE_CLOUD,
      region: values.region,
      cpu: values.google.cpu,
      ram: values.google.ram,
      hdd: values.google.hdd,
      'node-size': values.google.nodeSize,
      'node-count': values.nodeCount,
    },
  });
}

function trackApiProvision(
  provider: KaasProvider,
  values: FormValues,
  trackEvent: (action: string, properties: TrackEventProps) => void
) {
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider,
      region: values.region,
      'node-size': values.api.nodeSize,
      'node-count': values.nodeCount,
    },
  });
}

function trackAmazonProvision(
  values: FormValues,
  trackEvent: (action: string, properties: TrackEventProps) => void
) {
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

export function providerFormType(provider: KaasProvider) {
  switch (provider) {
    case KaasProvider.GOOGLE_CLOUD:
      return KaaSFormType.GKE;
    case KaasProvider.AWS:
      return KaaSFormType.EKS;
    case KaasProvider.AZURE:
      return KaaSFormType.AZURE;
    case KaasProvider.DIGITAL_OCEAN:
    case KaasProvider.LINODE:
    case KaasProvider.CIVO:
    default:
      return KaaSFormType.API;
  }
}
