import { Field, Form, useFormikContext } from 'formik';
import { useEffect, useMemo, useState } from 'react';

import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input, Select } from '@/portainer/components/form-components/Input';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';
import { Loading } from '@/portainer/components/widget/Loading';
import {
  Credential,
  KaasProvider,
  providerTitles,
} from '@/portainer/settings/cloud/types';
import { WarningAlert } from '@/portainer/components/Alert/WarningAlert';
import { Link } from '@/portainer/components/Link';
import { MetadataFieldset } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset';

import { useCloudProviderOptions } from '../queries';
import { useSetAvailableOption } from '../useSetAvailableOption';
import {
  CreateGKEClusterFormValues,
  isGKEKaasInfo,
  GKEKaasInfo,
} from '../types';
import { useIsKaasNameValid } from '../useIsKaasNameValid';

import { maxGKERam, minGKERam } from './validation';

type Props = {
  credentials: Credential[];
  provider: KaasProvider;
  name: string;
  setGKEKaasInfo: (info: GKEKaasInfo) => void;
  setvCPUCount: (vCPUCount: number) => void;
};

export function GKECreateClusterForm({
  credentials,
  provider,
  name,
  setGKEKaasInfo,
  setvCPUCount,
}: Props) {
  const { values, setFieldValue, errors, handleSubmit, isSubmitting, isValid } =
    useFormikContext<CreateGKEClusterFormValues>();
  const {
    region,
    credentialId,
    cpu,
    ram,
    kubernetesVersion,
    networkId,
    nodeSize,
  } = values;
  const [selectedCredential, setSelectedCredential] = useState<Credential>(
    credentials[0]
  );
  const cloudOptionsQuery = useCloudProviderOptions(
    selectedCredential,
    provider
  );
  const isNameValid = useIsKaasNameValid(name);

  const filteredNetworkOptions = useMemo(() => {
    const shortenedRegion = removeTextAfterLastHyphen(region);
    return cloudOptionsQuery?.data?.networks?.get(shortenedRegion) || [];
  }, [region, cloudOptionsQuery.data?.networks]);
  const credentialOptions = useMemo(
    () =>
      credentials.map((c) => ({
        value: c.id,
        label: c.name,
      })),
    [credentials]
  );
  const kubernetesVersions = useMemo(
    () => cloudOptionsQuery.data?.kubernetesVersions || [],
    [cloudOptionsQuery.data?.kubernetesVersions]
  );
  const regions = useMemo(
    () => cloudOptionsQuery.data?.regions || [],
    [cloudOptionsQuery.data?.regions]
  );
  const nodeSizes = useMemo(
    () => cloudOptionsQuery.data?.nodeSizes || [],
    [cloudOptionsQuery.data?.nodeSizes]
  );

  // if the selected credential id changes, update the credential
  useEffect(() => {
    setSelectedCredential(
      credentials.find((c) => c.id === Number(credentialId)) || credentials[0]
    );
  }, [credentialId, setSelectedCredential, credentials]);

  // if the vCPU count changes, update the vCPU count for validation
  useEffect(() => {
    setvCPUCount(cpu);
    // if the ram is out of the new valid range, change it
    if (ram < minGKERam(cpu)) {
      setFieldValue('ram', minGKERam(cpu));
    }
    if (ram > maxGKERam(cpu)) {
      setFieldValue('ram', maxGKERam(cpu));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cpu, setvCPUCount, setFieldValue]);

  // if the credentials change, select the first credential available
  useEffect(() => {
    const credential = credentials[0];
    setSelectedCredential(credential);
    setFieldValue('credentialId', credential.id);
  }, [credentials, setFieldValue]);

  // set each form value to a valid option when the options change
  useSetAvailableOption<string>(regions, region, 'region');
  useSetAvailableOption<string>(nodeSizes, nodeSize, 'nodeSize');
  useSetAvailableOption<string>(
    kubernetesVersions,
    kubernetesVersion,
    'kubernetesVersion'
  );
  useSetAvailableOption<string>(filteredNetworkOptions, networkId, 'networkId');

  // pass options to parent component to use for validation
  useEffect(() => {
    if (cloudOptionsQuery.data && isGKEKaasInfo(cloudOptionsQuery.data)) {
      setGKEKaasInfo(cloudOptionsQuery.data);
    }
  }, [cloudOptionsQuery.data, setFieldValue, setGKEKaasInfo]);

  return (
    <Form className="form-horizontal" onSubmit={handleSubmit} noValidate>
      <FormControl
        label="Credentials"
        tooltip="Credentials to create your cluster"
        inputId="kaas-credential"
        errors={errors.credentialId}
      >
        <Field
          name="credentialId"
          as={Select}
          type="number"
          id="kaa-credential"
          data-cy="kaasCreateForm-crdentialSelect"
          disabled={credentialOptions.length <= 1}
          options={credentialOptions}
        />
      </FormControl>
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
      {cloudOptionsQuery.data && isGKEKaasInfo(cloudOptionsQuery.data) && (
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
            errors={errors.nodeSize}
          >
            <Field
              name="nodeSize"
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
                errors={errors.cpu}
              >
                <Field
                  name="cpu"
                  as={Input}
                  type="number"
                  data-cy="kaasCreateForm-cpuInput"
                  step={2}
                  min={cloudOptionsQuery.data.cpu.min}
                  max={cloudOptionsQuery.data.cpu.max}
                  id="kaas-cpu"
                  placeholder={cloudOptionsQuery.data.cpu.default}
                />
              </FormControl>
              <FormControl
                label="Node RAM (GB)"
                tooltip="Amount of RAM (GB) in each node"
                inputId="kaas-nodeRAM"
                errors={errors.ram}
              >
                <Field
                  name="ram"
                  as={Input}
                  type="number"
                  data-cy="kaasCreateForm-ramInput"
                  step={1}
                  min={minGKERam(cpu)}
                  max={maxGKERam(cpu)}
                  id="kaas-ram"
                  placeholder={cloudOptionsQuery.data.cpu.default}
                />
              </FormControl>
            </>
          )}
          <FormControl
            label="Node disk space (GB)"
            tooltip="Amount of disk space (GB) in each node"
            inputId="kaas-nodeHDD"
            errors={errors.hdd}
          >
            <Field
              name="hdd"
              as={Input}
              type="number"
              data-cy="kaasCreateForm-hddInput"
              min={cloudOptionsQuery.data.hdd.min}
              max={cloudOptionsQuery.data.hdd.max}
              id="kaas-hdd"
              placeholder={cloudOptionsQuery.data.hdd.default}
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
            errors={errors.networkId}
          >
            <Field
              name="networkId"
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
      <MetadataFieldset />
      <FormSectionTitle>Actions</FormSectionTitle>
      <div className="form-group">
        <div className="col-sm-12">
          <LoadingButton
            disabled={!isValid || !isNameValid}
            isLoading={isSubmitting}
            loadingText="Provision in progress..."
          >
            <i className="fa fa-plus space-right" aria-hidden="true" />
            Provision environment
          </LoadingButton>
        </div>
      </div>
    </Form>
  );
}

function removeTextAfterLastHyphen(text: string) {
  const lastHyphenIndex = text.lastIndexOf('-');
  if (lastHyphenIndex > 0) {
    return text.substring(0, lastHyphenIndex);
  }
  return text;
}
