import { Field, useFormikContext } from 'formik';
import { useMemo, useState } from 'react';

import {
  Credential,
  KaasProvider,
  providerTitles,
} from '@/portainer/settings/cloud/types';

import { FormControl } from '@@/form-components/FormControl';
import { Input, Select } from '@@/form-components/Input';
import { Loading } from '@@/Widget/Loading';
import { Link } from '@@/Link';
import { TextTip } from '@@/Tip/TextTip';

import { CredentialsField } from '../shared/CredentialsField';
import { FormValues, InstanceTypeRegions, isEKSKaasInfo } from '../types';
import { useSetAvailableOption } from '../useSetAvailableOption';
import { useCloudProviderOptions } from '../queries';
import { ActionsSection } from '../shared/ActionsSection';
import { MoreSettingsSection } from '../../shared/MoreSettingsSection';
import { KaaSInfoText } from '../shared/KaaSInfoText';
import { NameField } from '../../shared/NameField';

type Props = {
  credentials: Credential[];
  provider: KaasProvider;
  isSubmitting: boolean;
};

// ApiCreateClusterForm handles form changes, conditionally renders inputs, and manually set values
export function EKSCreateClusterForm({
  credentials,
  provider,
  isSubmitting,
}: Props) {
  const { values, errors } = useFormikContext<FormValues>();
  const [isOptionsForce, setIsOptionsForce] = useState(false);

  const {
    credentialId,
    region,
    kubernetesVersion,
    amazon: { amiType, instanceType },
  } = values;

  const selectedCredential =
    credentials.find((c) => c.id === credentialId) || credentials[0];

  const cloudOptionsQuery = useCloudProviderOptions(
    provider,
    isEKSKaasInfo,
    selectedCredential,
    isOptionsForce
  );

  const cloudOptions = cloudOptionsQuery.data;

  const filteredInstanceTypes = useMemo(() => {
    if (cloudOptions) {
      return filterByAmiAndRegion(cloudOptions.instanceTypes, amiType, region);
    }
    return [];
  }, [region, amiType, cloudOptions]);

  const kubernetesVersions = useMemo(
    () => cloudOptions?.kubernetesVersions || [],
    [cloudOptions?.kubernetesVersions]
  );
  const regions = useMemo(
    () => cloudOptions?.regions || [],
    [cloudOptions?.regions]
  );
  const amiTypes = useMemo(() => {
    if (cloudOptions && isEKSKaasInfo(cloudOptions)) {
      return cloudOptions?.amiTypes || [];
    }
    return [];
  }, [cloudOptions]);

  const credentialOptions = useMemo(
    () =>
      credentials.map((c) => ({
        value: c.id,
        label: c.name,
      })),
    [credentials]
  );

  // ensure the form values are valid when the options change
  useSetAvailableOption(credentialOptions, credentialId, 'credentialId');
  useSetAvailableOption(regions, region, 'region');
  useSetAvailableOption(amiTypes, amiType, 'amazon.amiType');
  useSetAvailableOption(
    filteredInstanceTypes,
    instanceType,
    'amazon.instanceType'
  );
  useSetAvailableOption(
    kubernetesVersions,
    kubernetesVersion,
    'kubernetesVersion'
  );

  return (
    <>
      <KaaSInfoText />
      <NameField
        tooltip="Name of the cluster and environment"
        placeholder="e.g. my-cluster-name"
      />
      <CredentialsField credentials={credentials} />

      {cloudOptionsQuery.isError && (
        <TextTip color="orange">
          Error getting {providerTitles[provider]} info. Go to&nbsp;
          <Link
            to="portainer.settings.cloud.credential"
            params={{ id: credentialId }}
            title="cloud settings"
          >
            cloud settings
          </Link>
          &nbsp;to ensure the API key is valid.
        </TextTip>
      )}
      {cloudOptionsQuery.isLoading && <Loading />}
      {/* cluster details inputs */}
      {cloudOptions && (
        <>
          <FormControl
            label="Region"
            tooltip="Region in which to provision the cluster"
            inputId="kaas-region"
            errors={errors.region}
          >
            <Field
              name="region"
              as={Select}
              id="kaa-region"
              data-cy="kaasCreateForm-regionSelect"
              options={regions}
            />
          </FormControl>
          <FormControl
            label="AMI type"
            tooltip="Base image for Amazon EKS nodes"
            inputId="kaas-amiType"
            errors={errors.amazon?.amiType}
          >
            <Field
              name="amazon.amiType"
              as={Select}
              id="kaas-amiType"
              data-cy="kaasCreateForm-amiType"
              options={cloudOptions.amiTypes || []}
            />
          </FormControl>
          {region && (
            <FormControl
              label="Instance type"
              tooltip="Instance type of each node deployed in the cluster"
              inputId="kaas-instanceType"
              errors={errors.amazon?.instanceType}
            >
              <Field
                name="amazon.instanceType"
                as={Select}
                id="kaas-instanceType"
                data-cy="kaasCreateForm-instanceTypeSelect"
                options={filteredInstanceTypes}
              />
            </FormControl>
          )}

          <FormControl
            label="Node disk size (GiB)"
            tooltip="The size of the Elastic Block Storage (EBS) volume for each node"
            inputId="kaas-nodeVolumeSizeInput"
            errors={errors.amazon?.nodeVolumeSize}
          >
            <Field
              name="amazon.nodeVolumeSize"
              as={Input}
              type="number"
              data-cy="kaasCreateForm-nodeVolumeSizeInput"
              step="1"
              min={1}
              max={16000}
              id="kaas-nodeVolumeSize"
            />
          </FormControl>
          <FormControl
            label="Node count"
            tooltip="Number of nodes to provision in the cluster."
            inputId="kaas-nodeCount"
            errors={errors.nodeCount}
          >
            <Field
              name="nodeCount"
              as={Input}
              type="number"
              data-cy="kaasCreateForm-nodeCountInput"
              min="1"
              id="kaas-nodeCount"
              placeholder="3"
            />
          </FormControl>
          <FormControl
            label="Kubernetes version"
            tooltip="Kubernetes version running on the cluster"
            inputId="kaas-kubernetesVersion"
            errors={errors.kubernetesVersion}
          >
            <Field
              name="kubernetesVersion"
              as={Select}
              id="kaas-kubernetesVersion"
              data-cy="kaasCreateForm-kubernetesVersionSelect"
              options={kubernetesVersions}
            />
          </FormControl>
        </>
      )}

      <MoreSettingsSection>
        <TextTip color="blue">
          Metadata is only assigned to the environment in Portainer, i.e. the
          group and tags are not assigned to the cluster at the cloud provider
          end.
        </TextTip>
      </MoreSettingsSection>

      <ActionsSection
        onReloadClick={() => {
          setIsOptionsForce(true);
          cloudOptionsQuery.refetch();
        }}
        isReloading={
          cloudOptionsQuery.isLoading || cloudOptionsQuery.isFetching
        }
        isSubmitting={isSubmitting}
      />
    </>
  );
}

function filterByAmiAndRegion(
  instanceTypes: InstanceTypeRegions,
  amiType: string,
  region: string
) {
  if (amiType && region && instanceTypes[region]) {
    return (
      instanceTypes[region]
        .filter((i) => i.compatibleAmiTypes.includes(amiType))
        .map((i) => ({
          label: i.name,
          value: i.value,
        })) || []
    );
  }
  return [];
}
