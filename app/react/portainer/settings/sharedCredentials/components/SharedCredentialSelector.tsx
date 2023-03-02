import { Terminal } from 'lucide-react';

import Civo from '@/assets/ico/vendor/civo.svg?c';
import Linode from '@/assets/ico/vendor/linode.svg?c';
import Digitalocean from '@/assets/ico/vendor/digitalocean.svg?c';
import Googlecloud from '@/assets/ico/vendor/googlecloud.svg?c';
import Aws from '@/assets/ico/vendor/aws.svg?c';
import Azure from '@/assets/ico/vendor/azure.svg?c';

import { BoxSelector, BoxSelectorOption } from '@@/BoxSelector';
import { BadgeIcon } from '@@/BadgeIcon';

import { CredentialType } from '../types';

type Props = {
  value: CredentialType;
  onChange: (credentialType: CredentialType) => void;
};

const providerOptions: BoxSelectorOption<CredentialType>[] = [
  {
    id: CredentialType.CIVO,
    icon: Civo,
    label: 'Civo',
    description: 'Civo Kubernetes',
    value: CredentialType.CIVO,
  },
  {
    id: CredentialType.LINODE,
    icon: Linode,
    label: 'Linode',
    description: 'Linode Kubernetes Engine (LKE)',
    value: CredentialType.LINODE,
  },
  {
    id: CredentialType.DIGITAL_OCEAN,
    icon: Digitalocean,
    label: 'DigitalOcean',
    description: 'DigitalOcean Kubernetes (DOKS)',
    value: CredentialType.DIGITAL_OCEAN,
  },
  {
    id: CredentialType.GOOGLE_CLOUD,
    icon: Googlecloud,
    label: 'Google Cloud',
    description: 'Google Kubernetes Engine (GKE)',
    value: CredentialType.GOOGLE_CLOUD,
  },
  {
    id: CredentialType.AWS,
    icon: Aws,
    label: 'Amazon Web Services (AWS)',
    description: 'Elastic Kubernetes Service (EKS)',
    value: CredentialType.AWS,
  },
  {
    id: CredentialType.AZURE,
    icon: Azure,
    label: 'Microsoft Azure',
    description: 'Azure Kubernetes Service (AKS)',
    value: CredentialType.AZURE,
  },
  {
    id: CredentialType.SSH,
    icon: <BadgeIcon icon={Terminal} />,
    label: 'SSH',
    description:
      'Provision a Kubernetes cluster and install Portainer using SSH',
    value: CredentialType.SSH,
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
