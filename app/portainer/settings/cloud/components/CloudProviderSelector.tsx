import { BoxSelector } from '@/portainer/components/BoxSelector';
import { BoxSelectorOption } from '@/portainer/components/BoxSelector/types';

import { KaasProvider } from '../types';

type Props = {
  value: KaasProvider;
  onChange: (provider: KaasProvider) => void;
};

const providerOptions: BoxSelectorOption<KaasProvider>[] = [
  {
    id: KaasProvider.CIVO,
    icon: '',
    label: 'Civo',
    description: 'Civo Kubernetes',
    value: KaasProvider.CIVO,
  },
  {
    id: KaasProvider.LINODE,
    icon: 'fab fa-linode',
    label: 'Linode',
    description: 'Linode Kubernetes Engine (LKE)',
    value: KaasProvider.LINODE,
  },
  {
    id: KaasProvider.DIGITAL_OCEAN,
    icon: 'fab fa-digital-ocean',
    label: 'DigitalOcean',
    description: 'DigitalOcean Kubernetes (DOKS)',
    value: KaasProvider.DIGITAL_OCEAN,
  },
  {
    id: KaasProvider.GOOGLE_CLOUD,
    icon: 'fab fa-google',
    label: 'Google Cloud',
    description: 'Google Kubernetes Engine (GKE)',
    value: KaasProvider.GOOGLE_CLOUD,
  },
  {
    id: KaasProvider.AWS,
    icon: 'fab fa-aws',
    label: 'Amazon Web Services (AWS)',
    description: 'Elastic Kubernetes Service (EKS)',
    value: KaasProvider.AWS,
  },
  {
    id: KaasProvider.AZURE,
    icon: 'fab fa-microsoft',
    label: 'Microsoft Azure',
    description: 'Azure Kubernetes Service (AKS)',
    value: KaasProvider.AZURE,
  },
];

export function CloudProviderSelector({ value, onChange }: Props) {
  return (
    <BoxSelector
      radioName="cloudProvider"
      options={providerOptions}
      value={value}
      onChange={onChange}
    />
  );
}
