import { useMemo } from 'react';

import { KaasProvider } from '@/react/portainer/settings/cloud/types';
import Civo from '@/assets/ico/vendor/civo.svg?c';
import Linode from '@/assets/ico/vendor/linode.svg?c';
import Digitalocean from '@/assets/ico/vendor/digitalocean.svg?c';
import Googlecloud from '@/assets/ico/vendor/googlecloud.svg?c';
import Aws from '@/assets/ico/vendor/aws.svg?c';
import Azure from '@/assets/ico/vendor/azure.svg?c';
import Microk8s from '@/assets/ico/vendor/microk8s.svg?c';
import { notifyError } from '@/portainer/services/notifications';
import { usePublicSettings } from '@/react/portainer/settings/queries';

import { BoxSelector, BoxSelectorOption } from '@@/BoxSelector';
import { Loading } from '@@/Widget';

interface Props {
  provider: KaasProvider;
  onChange(value: KaasProvider): void;
}

const cloudProviderOptions: BoxSelectorOption<KaasProvider>[] = [
  {
    id: KaasProvider.CIVO,
    icon: Civo,
    label: 'Civo',
    description: 'Civo Kubernetes',
    value: KaasProvider.CIVO,
  },
  {
    id: KaasProvider.LINODE,
    icon: Linode,
    label: 'Linode',
    description: 'Linode Kubernetes Engine (LKE)',
    value: KaasProvider.LINODE,
  },
  {
    id: KaasProvider.DIGITAL_OCEAN,
    icon: Digitalocean,
    label: 'DigitalOcean',
    description: 'DigitalOcean Kubernetes (DOKS)',
    value: KaasProvider.DIGITAL_OCEAN,
  },
  {
    id: KaasProvider.GOOGLE_CLOUD,
    icon: Googlecloud,
    label: 'Google Cloud',
    description: 'Google Kubernetes Engine (GKE)',
    value: KaasProvider.GOOGLE_CLOUD,
  },
  {
    id: KaasProvider.AWS,
    icon: Aws,
    label: 'Amazon Web Services (AWS)',
    description: 'Elastic Kubernetes Service (EKS)',
    value: KaasProvider.AWS,
  },
  {
    id: KaasProvider.AZURE,
    icon: Azure,
    label: 'Microsoft Azure',
    description: 'Azure Kubernetes Service (AKS)',
    value: KaasProvider.AZURE,
  },
];

const microk8sOption = {
  id: KaasProvider.MICROK8S,
  icon: Microk8s,
  label: 'MicroK8s',
  description: 'Lightweight Kubernetes',
  value: KaasProvider.MICROK8S,
};

export function KaasProvidersSelector({ onChange, provider }: Props) {
  const { isLoading, data, isError } = usePublicSettings();

  const isMicrok8sEnabled = data?.Features.microk8s;

  const options = useMemo(() => {
    if (isMicrok8sEnabled) {
      return [...cloudProviderOptions, microk8sOption];
    }
    return cloudProviderOptions;
  }, [isMicrok8sEnabled]);

  if (isError) {
    notifyError('Unable to load public settings');
  }

  return (
    <>
      {isLoading && <Loading />}
      {!isLoading && (
        <BoxSelector
          radioName="kaas-type"
          data-cy="kaasCreateForm-providerSelect"
          options={options}
          onChange={(provider: KaasProvider) => onChange(provider)}
          value={provider}
        />
      )}
    </>
  );
}
