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
import { Option } from '@/portainer/components/form-components/Input/Select';
import { WarningAlert } from '@/portainer/components/Alert/WarningAlert';
import { Link } from '@/portainer/components/Link';
import { MetadataFieldset } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MetadataFieldset';

import { useCloudProviderOptions } from '../queries';
import { CreateApiClusterFormValues, isAPIKaasInfo } from '../types';
import { useIsKaasNameValid } from '../useIsKaasNameValid';
import { useSetAvailableOption } from '../useSetAvailableOption';

type Props = {
  credentials: Credential[];
  provider: KaasProvider;
  name: string;
};

// ApiCreateClusterForm handles form changes, conditionally renders inputs, and manually set values
export function ApiCreateClusterForm({ credentials, provider, name }: Props) {
  const { values, setFieldValue, errors, handleSubmit, isSubmitting, isValid } =
    useFormikContext<CreateApiClusterFormValues>();
  const { region, credentialId, networkId, nodeSize, kubernetesVersion } =
    values;
  const [selectedCredential, setSelectedCredential] = useState<Credential>(
    credentials[0]
  );
  const cloudOptionsQuery = useCloudProviderOptions(
    selectedCredential,
    provider
  );
  const isNameValid = useIsKaasNameValid(name);

  const filteredNetworkOptions = useMemo(
    () => cloudOptionsQuery?.data?.networks?.get(region) || [],
    [region, cloudOptionsQuery.data]
  );

  // if the selected credential id changes, update the credential
  useEffect(() => {
    setSelectedCredential(
      credentials.find((c) => c.id === Number(credentialId)) || credentials[0]
    );
  }, [credentialId, setSelectedCredential, credentials]);

  // if the credentials change, select the first credential available
  useEffect(() => {
    const credential = credentials[0];
    setSelectedCredential(credential);
    setFieldValue('credentialId', credential.id);
  }, [credentials, setFieldValue]);

  const credentialOptions: Option<number>[] = credentials.map((c) => ({
    value: c.id,
    label: c.name,
  }));

  // when the options change, set the available options in the select inputs
  useSetAvailableOption(filteredNetworkOptions, networkId || '', 'networkId');
  useSetAvailableOption(cloudOptionsQuery.data?.regions, region, 'region');
  useSetAvailableOption(
    cloudOptionsQuery.data?.nodeSizes,
    nodeSize,
    'nodeSize'
  );
  useSetAvailableOption(
    cloudOptionsQuery.data?.kubernetesVersions,
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
          id="kaa-credential"
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
      {cloudOptionsQuery.data && isAPIKaasInfo(cloudOptionsQuery.data) && (
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
              options={cloudOptionsQuery.data.regions}
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
              options={cloudOptionsQuery.data?.nodeSizes}
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
              min={1}
              max={1000}
              id="kaas-nodeCount"
              placeholder="3"
            />
          </FormControl>
          {region && filteredNetworkOptions.length > 0 && (
            <FormControl
              label="Network ID"
              tooltip="ID of network attached to the cluster"
              inputId="kaas-networkId"
              errors={errors.networkId}
            >
              <Field
                name="networkId"
                as={Select}
                id="kaas-networkId"
                data-cy="kaasCreateForm-networkIdSelect"
                disabled={filteredNetworkOptions.length === 1}
                options={filteredNetworkOptions}
              />
            </FormControl>
          )}
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
              options={cloudOptionsQuery.data?.kubernetesVersions}
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
