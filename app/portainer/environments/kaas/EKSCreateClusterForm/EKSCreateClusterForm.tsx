import { Field, Form, useFormikContext } from 'formik';
import { useEffect, useMemo, useState } from 'react';

import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input, Select } from '@/portainer/components/form-components/Input';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';
import { Loading } from '@/portainer/components/widget/Loading';
import { MetadataFieldset } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset';
import {
  Credential,
  KaasProvider,
  providerTitles,
} from '@/portainer/settings/cloud/types';
import { WarningAlert } from '@/portainer/components/Alert/WarningAlert';
import { Link } from '@/portainer/components/Link';

import { useCloudProviderOptions } from '../queries';
import { useSetAvailableOption } from '../useSetAvailableOption';
import {
  CreateEKSClusterFormValues,
  InstanceTypeRegions,
  isEKSKaasInfo,
} from '../types';
import { useIsKaasNameValid } from '../useIsKaasNameValid';

type Props = {
  credentials: Credential[];
  provider: KaasProvider;
  name: string;
};

// ApiCreateClusterForm handles form changes, conditionally renders inputs, and manually set values
export function EKSCreateClusterForm({ credentials, provider, name }: Props) {
  const { values, errors, handleSubmit, isSubmitting, isValid } =
    useFormikContext<CreateEKSClusterFormValues>();
  const { credentialId, region, amiType, instanceType, kubernetesVersion } =
    values;
  const [selectedCredential, setSelectedCredential] = useState<Credential>(
    credentials[0]
  );
  const cloudOptionsQuery = useCloudProviderOptions(
    selectedCredential,
    provider
  );
  const isNameValid = useIsKaasNameValid(name);

  const filteredInstanceTypes = useMemo(() => {
    if (cloudOptionsQuery.data && isEKSKaasInfo(cloudOptionsQuery.data)) {
      return filterByAmiAndRegion(
        cloudOptionsQuery.data.instanceTypes,
        amiType,
        region
      );
    }
    return [];
  }, [region, amiType, cloudOptionsQuery.data]);
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
  const amiTypes = useMemo(() => {
    if (cloudOptionsQuery.data && isEKSKaasInfo(cloudOptionsQuery.data)) {
      return cloudOptionsQuery.data?.amiTypes || [];
    }
    return [];
  }, [cloudOptionsQuery.data]);

  // if the selected credential id changes, update the credential
  useEffect(() => {
    setSelectedCredential(
      credentials.find((c) => c.id === Number(credentialId)) || credentials[0]
    );
  }, [credentialId, setSelectedCredential, credentials]);

  // ensure the form values are valid when the options change
  useSetAvailableOption(credentialOptions, credentialId, 'credentialId');
  useSetAvailableOption(regions, region, 'region');
  useSetAvailableOption(amiTypes, amiType, 'amiType');
  useSetAvailableOption(filteredInstanceTypes, instanceType, 'instanceType');
  useSetAvailableOption(
    kubernetesVersions,
    kubernetesVersion,
    'kubernetesVersion'
  );

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
          id="kaas-credential"
          data-cy="kaasCreateForm-crdentialSelect"
          disabled={credentialOptions.length === 1}
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
      {cloudOptionsQuery.data && isEKSKaasInfo(cloudOptionsQuery.data) && (
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
            errors={errors.amiType}
          >
            <Field
              name="amiType"
              as={Select}
              id="kaas-amiType"
              data-cy="kaasCreateForm-amiType"
              options={cloudOptionsQuery.data?.amiTypes || []}
            />
          </FormControl>
          {region && (
            <FormControl
              label="Instance type"
              tooltip="Instance type of each node deployed in the cluster"
              inputId="kaas-instanceType"
              errors={errors.instanceType}
            >
              <Field
                name="instanceType"
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
            errors={errors.nodeVolumeSize}
          >
            <Field
              name="nodeVolumeSize"
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
