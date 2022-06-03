import { BoxSelector, buildOption } from '@/portainer/components/BoxSelector';
import { BoxSelectorOption } from '@/portainer/components/BoxSelector/types';
import { KaasProvider } from '@/portainer/settings/cloud/types';

const linodeOptions = buildOption(
  KaasProvider.LINODE,
  'fab fa-linode',
  'Linode',
  'Linode Kubernetes Engine (LKE)',
  KaasProvider.LINODE
);

const digitalOceanOptions = buildOption(
  KaasProvider.DIGITAL_OCEAN,
  'fab fa-digital-ocean',
  'DigitalOcean',
  'DigitalOcean Kubernetes (DOKS)',
  KaasProvider.DIGITAL_OCEAN
);

const civoOptions = buildOption(
  KaasProvider.CIVO,
  'fab fa-civo',
  'Civo',
  'Civo Kubernetes',
  KaasProvider.CIVO
);

const azureOptions = buildOption(
  KaasProvider.AZURE,
  'fab fa-microsoft',
  'Azure',
  'Azure Kubernetes Service (AKS)',
  KaasProvider.AZURE
);

const gkeOptions = buildOption(
  KaasProvider.GOOGLE_CLOUD,
  'fab fa-google',
  'Google Cloud',
  'Google Kubernetes Engine (GKE)',
  KaasProvider.GOOGLE_CLOUD
);

const awsOptions = buildOption(
  KaasProvider.AWS,
  'fab fa-aws',
  'AWS',
  'Elastic Kubernetes Service (EKS)',
  KaasProvider.AWS
);

const boxSelectorOptions: BoxSelectorOption<
  | KaasProvider.CIVO
  | KaasProvider.LINODE
  | KaasProvider.DIGITAL_OCEAN
  | KaasProvider.AZURE
  | KaasProvider.GOOGLE_CLOUD
  | KaasProvider.AWS
>[] = [
  civoOptions,
  linodeOptions,
  digitalOceanOptions,
  azureOptions,
  gkeOptions,
  awsOptions,
];

interface Props {
  provider: KaasProvider;
  onChange(value: KaasProvider): void;
}

export function KaasProvidersSelector({ onChange, provider }: Props) {
  return (
    <BoxSelector
      radioName="kaas-type"
      data-cy="kaasCreateForm-providerSelect"
      options={boxSelectorOptions}
      onChange={(provider: KaasProvider) => onChange(provider)}
      value={provider}
    />
  );
}
