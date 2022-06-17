import { Field, useFormikContext } from 'formik';
import { useMemo, useState } from 'react';

import {
  Credential,
  KaasProvider,
  providerTitles,
} from '@/portainer/settings/cloud/types';
import { MoreSettingsSection } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/MoreSettingsSection';
import { NameField } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/shared/NameField';

import { Select as ReactSelect } from '@@/form-components/ReactSelect';
import { FormControl } from '@@/form-components/FormControl';
import { Input, Select } from '@@/form-components/Input';
import { Loading } from '@@/Widget/Loading';
import { Option } from '@@/form-components/Input/Select';
import { Link } from '@@/Link';
import { TextTip } from '@@/Tip/TextTip';

import { useCloudProviderOptions } from '../queries';
import { FormValues, isAzureKaasInfo } from '../types';
import { useSetAvailableOption } from '../useSetAvailableOption';
import { CredentialsField } from '../shared/CredentialsField';
import { ActionsSection } from '../shared/ActionsSection';
import { KaasInfoText } from '../shared/KaasInfoText';

type Props = {
  credentials: Credential[];
  provider: KaasProvider;
  isSubmitting: boolean;
};

// ApiCreateClusterForm handles form changes, conditionally renders inputs, and manually set values
export function AzureCreateClusterForm({
  credentials,
  provider,
  isSubmitting,
}: Props) {
  const { values, errors, setFieldValue } = useFormikContext<FormValues>();
  const {
    region,
    credentialId,
    kubernetesVersion,
    azure: {
      resourceGroup,
      tier,
      availabilityZones,
      resourceGroupInput,
      nodeSize,
    },
  } = values;
  const [isOptionsForce, setIsOptionsForce] = useState(false);

  const selectedCredential =
    credentials.find((c) => c.id === credentialId) || credentials[0];

  const cloudOptionsQuery = useCloudProviderOptions(
    provider,
    isAzureKaasInfo,
    selectedCredential,
    isOptionsForce
  );

  // update the node size options based on the selected region
  const filteredNodeSizes = useMemo(() => {
    if (cloudOptionsQuery.data?.nodeSizes[region]) {
      return (
        cloudOptionsQuery.data?.nodeSizes[region].map((ns) => ({
          label: ns.name,
          value: ns.value,
        })) || []
      );
    }
    return [];
  }, [region, cloudOptionsQuery.data?.nodeSizes]);
  // update the availabilityZoneOptions based on the node size selected inside the region
  const availabilityZoneOptions = useMemo(() => {
    if (nodeSize) {
      return (
        cloudOptionsQuery.data?.nodeSizes[region]
          ?.find((ns) => ns.value === nodeSize)
          ?.zones?.map((o) => ({ value: o, label: `Zone ${o}` })) || []
      );
    }
    return [];
  }, [nodeSize, region, cloudOptionsQuery.data?.nodeSizes]);
  // update the tier option label based on availability zones
  const tiers = useMemo(() => {
    const paidLabel =
      availabilityZones && availabilityZones.length ? '99.95%' : '99.9%';
    return (
      cloudOptionsQuery.data?.tiers.map((tier) => {
        if (tier === 'Paid') {
          return {
            label: `${paidLabel} - charges apply`,
            value: tier,
          };
        }
        return {
          label: '99.5% - free',
          value: tier,
        };
      }) || []
    );
  }, [cloudOptionsQuery.data?.tiers, availabilityZones]);
  const { regions, kubernetesVersions, resourceGroups } =
    cloudOptionsQuery.data || {
      regions: [],
      kubernetesVersions: [],
      resourceGroups: [],
    };

  // when the options change, set the available options in the select inputs
  useSetAvailableOption(
    resourceGroups,
    resourceGroup,
    'azure.resourceGroup',
    'None available'
  );
  useSetAvailableOption(tiers, tier, 'azure.tier');
  useSetAvailableOption(regions, region, 'region');
  useSetAvailableOption(
    kubernetesVersions,
    kubernetesVersion,
    'kubernetesVersion'
  );
  useSetAvailableOption(filteredNodeSizes, nodeSize, 'azure.nodeSize');

  return (
    <>
      <KaasInfoText />
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
      {cloudOptionsQuery.data &&
        !cloudOptionsQuery.isError &&
        isAzureKaasInfo(cloudOptionsQuery.data) && (
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
                id="kaas-region"
                data-cy="kaasCreateForm-regionSelect"
                disabled={regions.length <= 1}
                options={regions}
              />
            </FormControl>
            <FormControl
              label="Resource group"
              tooltip="A collection of Azure resources"
              inputId="kaas-resourceGroup"
              classes="!mb-0"
            >
              <div className="flex pb-0">
                <label
                  className="inline-flex items-center pt-3"
                  htmlFor="radioSelect"
                >
                  <Field
                    type="radio"
                    name="azure.resourceGroupInput"
                    id="radioSelect"
                    className="form-radio !mt-0"
                    value="select"
                  />
                  <span className="ml-4 text-sm font-normal">
                    Select existing resource group
                  </span>
                </label>
                <label
                  className="inline-flex items-center ml-12 pt-3"
                  htmlFor="radioInput"
                >
                  <Field
                    type="radio"
                    name="azure.resourceGroupInput"
                    id="radioInput"
                    className="form-radio !mt-0"
                    value="input"
                  />
                  <span className="ml-4 text-sm font-normal">
                    Add a new resource group
                  </span>
                </label>
              </div>
            </FormControl>
            {/* Choose a resource group */}
            {resourceGroupInput === 'select' && (
              <FormControl
                label=""
                inputId="kaas-resourceGroup"
                classes="!mt-2"
                errors={
                  !resourceGroup
                    ? 'No resource groups available, please add a resource group.'
                    : ''
                }
              >
                <Field
                  name="azure.resourceGroup"
                  as={Select}
                  id="kaas-resourceGroup"
                  data-cy="kaasCreateForm-resourceGroup"
                  disabled={resourceGroups.length <= 1}
                  options={resourceGroups}
                />
              </FormControl>
            )}
            {/* or create a resource group */}
            {resourceGroupInput === 'input' && (
              <FormControl
                label=""
                inputId="kaas-resourceGroupName"
                errors={errors.azure?.resourceGroupName}
                classes="!mt-2"
              >
                <Field
                  name="azure.resourceGroupName"
                  as={Input}
                  id="kaas-resourceGroupName"
                  placeholder="e.g my-resource-group"
                  data-cy="kaasCreateForm-resourceGroupName"
                />
              </FormControl>
            )}
            <FormControl
              label="Node pool name"
              tooltip="Name of the node pool(s) within the cluster"
              inputId="kaas-name"
              errors={errors.azure?.poolName}
            >
              <Field
                name="azure.poolName"
                as={Input}
                id="kaas-poolName"
                data-cy="kaasCreateForm-poolNameInput"
                placeholder="e.g. pool1"
              />
            </FormControl>
            <FormControl
              label="Node size"
              tooltip="Size of each node deployed in the cluster. Check your Azure compute quota to ensure you have enough resources to deploy this cluster."
              inputId="kaas-nodeSize"
              errors={errors.azure?.nodeSize}
            >
              <Field
                name="azure.nodeSize"
                as={Select}
                id="kaas-nodeSize"
                data-cy="kaasCreateForm-nodeSizeSelect"
                options={filteredNodeSizes}
              />
            </FormControl>
            <TextTip color="blue">
              Check your
              <a
                href="https://portal.azure.com/#blade/Microsoft_Azure_Capacity/QuotaMenuBlade/myQuotas"
                target="_blank"
                rel="noopener noreferrer"
                className="mx-1"
              >
                Azure compute quota
              </a>
              to ensure you have enough resources to deploy this cluster
            </TextTip>
            <FormControl
              label="Node count"
              tooltip="Number of nodes to provision in the cluster. Check your Azure compute quota to ensure you have enough resources to deploy this cluster."
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
            {/* availability zones */}
            <FormControl
              label="Availability zones"
              tooltip="A high-availability offering that allows you to spread the nodes in this node pool across multiple physical locations, protecting your applications from datacenter failures"
              inputId="kaas-availabilityZoneOptions"
              errors={errors.azure?.availabilityZones}
            >
              {availabilityZoneOptions.length === 0 && (
                <Field
                  name="availabilityZonesNoneAvailable"
                  as={ReactSelect}
                  id="kaas-availabilityZonesNoneAvailable"
                  data-cy="kaasCreateForm-availabilityZonesNoneAvailable"
                  isDisabled
                  placeholder="None available"
                  options={[{ label: 'None available', value: '' }]}
                />
              )}
              {availabilityZoneOptions.length > 0 && (
                <Field
                  name="availabilityZones"
                  as={ReactSelect}
                  isMulti
                  closeMenuOnSelect={false}
                  value={
                    availabilityZones
                      ? availabilityZones.find((o) =>
                          availabilityZoneOptions?.includes({
                            value: o,
                            label: o,
                          })
                        )
                      : ''
                  }
                  onChange={(options: Option<string>[]) =>
                    setFieldValue(
                      'azure.availabilityZones',
                      options.map((o) => o.value)
                    )
                  }
                  options={availabilityZoneOptions}
                  id="kaas-availabilityZoneOptions"
                  data-cy="kaasCreateForm-availabilityZoneOptions"
                />
              )}
            </FormControl>
            {/* tier */}
            <FormControl
              label="API server availability"
              tooltip="The uptime service level agreement that guarantees a Kubernetes API server uptime of 99.95% for clusters with one or more availability zones and 99.9% for all other clusters"
              inputId="kaas-tier"
              errors={errors.azure?.tier}
            >
              <Field
                name="azure.tier"
                as={Select}
                id="kaas-tier"
                data-cy="kaasCreateForm-tierSelect"
                disabled={tiers.length <= 1}
                options={tiers}
              />
            </FormControl>
            {/* dns prefix */}
            <FormControl
              label="DNS name prefix"
              tooltip="DNS name prefix to use with the hosted Kubernetes API server FQDN. You will use this to connect to the Kubernetes API when managing containers after creating the cluster."
              inputId="kaas-dnsPrefix"
              errors={errors.azure?.dnsPrefix}
            >
              <Field
                name="azure.dnsPrefix"
                as={Input}
                data-cy="kaasCreateForm-dnsPrefix"
                id="kaas-dnsPrefix"
                placeholder="e.g. cluster1"
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
                disabled={kubernetesVersions.length <= 1}
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
