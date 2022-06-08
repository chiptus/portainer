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
import { Alert } from '@/portainer/components/Alert/Alert';
import { Link } from '@/portainer/components/Link';
import { TextTip } from '@/portainer/components/Tip/TextTip';
import { MoreSettingsSection } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MoreSettingsSection';

import { useCloudProviderOptions } from '../queries';
import { FormValues, isAPIKaasInfo } from '../types';
import { useSetAvailableOption } from '../useSetAvailableOption';
import { CredentialsField } from '../shared/CredentialsField';
import { ActionsSection } from '../shared/ActionsSection';

type Props = {
  credentials: Credential[];
  provider: KaasProvider;
  isSubmitting: boolean;
};

// ApiCreateClusterForm handles form changes, conditionally renders inputs, and manually set values
export function ApiCreateClusterForm({
  credentials,
  provider,
  isSubmitting,
}: Props) {
  const [isOptionsForce, setIsOptionsForce] = useState(false);
  const { values, setFieldValue, errors } = useFormikContext<FormValues>();
  const {
    region,
    credentialId,
    nodeSize,
    kubernetesVersion,
    api: { networkId },
  } = values;

  const selectedCredential =
    credentials.find((c) => c.id === credentialId) || credentials[0];

  const cloudOptionsQuery = useCloudProviderOptions(
    provider,
    isAPIKaasInfo,
    selectedCredential,
    isOptionsForce
  );

  const cloudOptions = cloudOptionsQuery.data;

  const filteredNetworkOptions = useMemo(
    () => cloudOptions?.networks?.[region] || [],
    [cloudOptions?.networks, region]
  );

  // if the credentials change, select the first credential available
  useEffect(() => {
    const credential = credentials[0];
    setFieldValue('credentialId', credential.id);
  }, [credentials, setFieldValue]);

  // when the options change, set the available options in the select inputs
  useSetAvailableOption(filteredNetworkOptions, networkId, 'api.networkId');
  useSetAvailableOption(cloudOptions?.regions, region, 'region');
  useSetAvailableOption(cloudOptions?.nodeSizes, nodeSize, 'nodeSize');
  useSetAvailableOption(
    cloudOptions?.kubernetesVersions,
    kubernetesVersion,
    'kubernetesVersion'
  );

  // when the region changes, update the selected network
  useEffect(() => {
    if (filteredNetworkOptions.length > 0 && region) {
      setFieldValue('api.networkId', filteredNetworkOptions[0].value);
    }
  }, [region, filteredNetworkOptions, setFieldValue]);

  return (
    <>
      <CredentialsField credentials={credentials} />
      {cloudOptionsQuery.isError && (
        <Alert>
          Error getting {providerTitles[provider]} info. Go to&nbsp;
          <Link
            to="portainer.settings.cloud.credential"
            params={{ id: credentialId }}
            title="cloud settings"
          >
            cloud settings
          </Link>
          &nbsp;to ensure the API key is valid.
        </Alert>
      )}
      {(cloudOptionsQuery.isLoading || cloudOptionsQuery.isFetching) && (
        <Loading />
      )}
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
              options={cloudOptions.regions}
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
              options={cloudOptions.nodeSizes}
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
          {region && filteredNetworkOptions.length > 0 && (
            <FormControl
              label="Network ID"
              tooltip="ID of network attached to the cluster"
              inputId="kaas-networkId"
              errors={errors.api?.networkId}
            >
              <Field
                name="api.networkId"
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
              options={cloudOptions.kubernetesVersions}
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
