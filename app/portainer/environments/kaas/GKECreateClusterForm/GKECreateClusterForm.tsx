import { Field, useFormikContext } from 'formik';
import { useEffect, useMemo, useState } from 'react';

import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input, Select } from '@/portainer/components/form-components/Input';
import { Loading } from '@/portainer/components/widget/Loading';
import {
  Credential,
  KaasProvider,
  providerTitles,
} from '@/portainer/settings/cloud/types';
import { WarningAlert } from '@/portainer/components/Alert/WarningAlert';
import { Link } from '@/portainer/components/Link';
import { TextTip } from '@/portainer/components/Tip/TextTip';
import { MoreSettingsSection } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MoreSettingsSection';
import { NameField } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';

import { useCloudProviderOptions } from '../queries';
import { useSetAvailableOption } from '../useSetAvailableOption';
import { FormValues, isGKEKaasInfo } from '../types';
import { CredentialsField } from '../shared/CredentialsField';
import { ActionsSection } from '../shared/ActionsSection';
import { KaasInfoText } from '../shared/KaasInfoText';

import { maxGKERam, minGKERam } from './validation';

type Props = {
  credentials: Credential[];
  provider: KaasProvider;
  isSubmitting: boolean;
};

export function GKECreateClusterForm({
  credentials,
  provider,
  isSubmitting,
}: Props) {
  const { values, setFieldValue, errors } = useFormikContext<FormValues>();
  const {
    region,
    credentialId,
    kubernetesVersion,
    google: { networkId, cpu, ram, nodeSize },
  } = values;
  const [isOptionsForce, setIsOptionsForce] = useState(false);

  const selectedCredential =
    credentials.find((c) => c.id === credentialId) || credentials[0];

  const cloudOptionsQuery = useCloudProviderOptions(
    provider,
    isGKEKaasInfo,
    selectedCredential,
    isOptionsForce
  );

  const cloudOptions = cloudOptionsQuery.data;

  const filteredNetworkOptions = useMemo(() => {
    const shortenedRegion = removeTextAfterLastHyphen(region);
    return cloudOptions?.networks?.[shortenedRegion] || [];
  }, [region, cloudOptions?.networks]);

  const kubernetesVersions = useMemo(
    () => cloudOptions?.kubernetesVersions || [],
    [cloudOptions?.kubernetesVersions]
  );
  const regions = useMemo(
    () => cloudOptions?.regions || [],
    [cloudOptions?.regions]
  );
  const nodeSizes = useMemo(
    () => cloudOptions?.nodeSizes || [],
    [cloudOptions?.nodeSizes]
  );

  // if the vCPU count changes, update the vCPU count for validation
  useEffect(() => {
    // if the ram is out of the new valid range, change it
    if (ram < minGKERam(cpu)) {
      setFieldValue('google.ram', minGKERam(cpu));
    }
    if (ram > maxGKERam(cpu)) {
      setFieldValue('google.ram', maxGKERam(cpu));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cpu, setFieldValue]);

  // if the credentials change, select the first credential available
  useEffect(() => {
    if (credentials.length > 0) {
      const credential = credentials[0];

      setFieldValue('credentialId', credential.id);
    }
  }, [credentials, setFieldValue]);
  // set each form value to a valid option when the options change
  useSetAvailableOption(regions, region, 'region');
  useSetAvailableOption(nodeSizes, nodeSize, 'google.nodeSize');
  useSetAvailableOption(
    kubernetesVersions,
    kubernetesVersion,
    'kubernetesVersion'
  );
  useSetAvailableOption(filteredNetworkOptions, networkId, 'google.networkId');

  return (
    <>
      <KaasInfoText />
      <NameField
        tooltip="Name of the cluster and environment"
        placeholder="e.g. my-cluster-name"
      />
      <CredentialsField credentials={credentials} />

      {cloudOptionsQuery.isError && (
        <WarningAlert>
          Error getting {providerTitles[provider]} info. Go to&nbsp;
          <Link
            to="portainer.settings.cloud.credential"
            params={{ id: credentialId }}
            title="cloud settings"
          >
            cloud settings
          </Link>
          &nbsp;to ensure the API key is valid.
        </WarningAlert>
      )}
      {cloudOptionsQuery.isLoading && <Loading />}
      {/* cluster details inputs */}
      {cloudOptions && isGKEKaasInfo(cloudOptions) && (
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
              disabled={regions.length <= 1}
            />
          </FormControl>
          <FormControl
            label="Node size"
            tooltip="Size of each node deployed in the cluster"
            inputId="kaas-nodeSize"
            errors={errors.google?.nodeSize}
          >
            <Field
              name="google.nodeSize"
              as={Select}
              id="kaas-nodeSize"
              data-cy="kaasCreateForm-nodeSizeSelect"
              options={nodeSizes}
              disabled={nodeSizes.length <= 1}
            />
          </FormControl>
          {nodeSize === 'custom' && (
            <>
              <FormControl
                label="Node vCPUs"
                tooltip="Number of vCPU cores in each node"
                inputId="kaas-nodeCPU"
                errors={errors.google?.cpu}
              >
                <Field
                  name="google.cpu"
                  as={Input}
                  type="number"
                  data-cy="kaasCreateForm-cpuInput"
                  step={2}
                  min={cloudOptions.cpu.min}
                  max={cloudOptions.cpu.max}
                  id="kaas-cpu"
                  placeholder={cloudOptions.cpu.default}
                />
              </FormControl>
              <FormControl
                label="Node RAM (GB)"
                tooltip="Amount of RAM (GB) in each node"
                inputId="kaas-nodeRAM"
                errors={errors.google?.ram}
              >
                <Field
                  name="google.ram"
                  as={Input}
                  type="number"
                  data-cy="kaasCreateForm-ramInput"
                  step={1}
                  min={minGKERam(cpu)}
                  max={maxGKERam(cpu)}
                  id="kaas-ram"
                  placeholder={cloudOptions.cpu.default}
                />
              </FormControl>
            </>
          )}
          <FormControl
            label="Node disk space (GB)"
            tooltip="Amount of disk space (GB) in each node"
            inputId="kaas-nodeHDD"
            errors={errors.google?.hdd}
          >
            <Field
              name="google.hdd"
              as={Input}
              type="number"
              data-cy="kaasCreateForm-hddInput"
              min={cloudOptions.hdd.min}
              max={cloudOptions.hdd.max}
              id="kaas-hdd"
              placeholder={cloudOptions.hdd.default}
            />
          </FormControl>
          <FormControl
            label="Node count"
            tooltip="Number of nodes to provision in the cluster"
            inputId="kaas-nodeCount"
            errors={errors.nodeCount}
          >
            <Field
              name="nodeCount"
              as={Input}
              type="number"
              data-cy="kaasCreateForm-nodeCountInput"
              min={1}
              max={1000}
              id="kaas-nodeCount"
              placeholder="3"
            />
          </FormControl>
          <FormControl
            label="Subnet"
            tooltip="Name of the subnet attached to the cluster"
            inputId="kaas-networkId"
            errors={errors.google?.networkId}
          >
            <Field
              name="google.networkId"
              as={Select}
              id="kaas-networkId"
              data-cy="kaasCreateForm-networkIdSelect"
              options={filteredNetworkOptions}
              disabled={filteredNetworkOptions.length <= 1}
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
              disabled={kubernetesVersions.length <= 1}
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

function removeTextAfterLastHyphen(text: string) {
  const lastHyphenIndex = text.lastIndexOf('-');
  if (lastHyphenIndex > 0) {
    return text.substring(0, lastHyphenIndex);
  }
  return text;
}
