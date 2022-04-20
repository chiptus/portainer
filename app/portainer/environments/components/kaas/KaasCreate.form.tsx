import { Field, Form, Formik, useFormikContext } from 'formik';
import { useRouter } from '@uirouter/react';
import { useEffect, useState } from 'react';

import { react2angular } from '@/react-tools/react2angular';
import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input, Select } from '@/portainer/components/form-components/Input';
import { BoxSelector, buildOption } from '@/portainer/components/BoxSelector';
import { BoxSelectorOption } from '@/portainer/components/BoxSelector/types';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';
import { useSettings } from '@/portainer/settings/queries';
import { trackEvent } from '@/angulartics.matomo/analytics-services';
import { Loading } from '@/portainer/components/widget/Loading';
import { Link } from '@/portainer/components/Link';
import { CloudSettingsForm } from '@/portainer/settings/cloud/CloudSettingsForm';

import { useEnvironmentList } from '../../queries';

import { useCloudProviderOptions, useCreateKaasCluster } from './queries';
import {
  KaasCreateFormInitialValues,
  KaasCreateFormValues,
  KaasProvider,
  CloudApiKeys,
} from './kaas.types';
import { validationSchema } from './KaasCreate.form.validation';

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

const defaultValues = {
  name: KaasCreateFormInitialValues.name,
  nodeCount: KaasCreateFormInitialValues.nodeCount,
  region: '',
  nodeSize: '',
  kubernetesVersion: '',
};

const boxSelectorOptions: BoxSelectorOption<
  KaasProvider.CIVO | KaasProvider.LINODE | KaasProvider.DIGITAL_OCEAN
>[] = [civoOptions, linodeOptions, digitalOceanOptions];

const providerTitles = {
  civo: 'Civo',
  linode: 'Linode',
  digitalocean: 'DigitalOcean',
};

type Props = {
  view?: 'addEnvironment' | 'wizard';
  showTitle?: boolean;
  onUpdate?: () => void;
  onAnalytics?: (eventName: string) => void;
};

export function KaasCreateForm({
  view = 'addEnvironment',
  showTitle,
  onUpdate,
  onAnalytics,
}: Props) {
  const router = useRouter();

  const settingsQuery = useSettings();
  const [availableProviders, setavailableProviders] = useState<KaasProvider[]>(
    []
  );
  const cloudOptionsQuery = useCloudProviderOptions(
    availableProviders[0],
    !!availableProviders[0]
  );

  const environmentsQuery = useEnvironmentList();
  const environmentNames = environmentsQuery.environments?.map(
    (env) => env.Name
  );
  const createKaasCluster = useCreateKaasCluster();

  const [initialValues, setInitialValues] = useState<
    KaasCreateFormValues | undefined
  >(undefined);
  const [apiKeyField, setApiKeyField] = useState<string>('');
  // remember some form values
  const [initialName, setinitialName] = useState<string>(
    KaasCreateFormInitialValues.name
  );
  const [initialType, setInitialType] = useState<KaasProvider>(
    availableProviders[0] || KaasProvider.CIVO
  );

  // when the api keys change, update the available providers
  useEffect(() => {
    if (settingsQuery.data?.CloudApiKeys) {
      setavailableProviders(
        getAvailableProviders(settingsQuery.data.CloudApiKeys)
      );
    }
  }, [settingsQuery.data?.CloudApiKeys]);

  // set the initial form values when they are available
  useEffect(() => {
    if (cloudOptionsQuery.data && availableProviders[0]) {
      // only set the initial values once
      if (!initialValues) {
        const defaultRegion = cloudOptionsQuery.data?.regions[0].value;
        setInitialValues({
          name: initialName,
          nodeCount: KaasCreateFormInitialValues.nodeCount,
          type: initialType,
          kubernetesVersion: cloudOptionsQuery.data?.kubernetesVersions[0],
          nodeSize: cloudOptionsQuery.data?.nodeSizes[0].value,
          region: defaultRegion,
          networkId: cloudOptionsQuery.data?.networks?.get(defaultRegion)?.at(0)
            ?.value,
        });
      }
    }
  }, [cloudOptionsQuery.data, availableProviders]);

  // handle the submit with current form values
  async function onSubmit(formValues: KaasCreateFormValues) {
    if (settingsQuery.data?.EnableTelemetry) {
      sendKaasProvisionAnalytics(formValues);
    }
    createKaasCluster.mutate(formValues, {
      onSuccess: () => {
        if (onUpdate) {
          onUpdate();
        }
        if (view === 'addEnvironment') {
          router.stateService.go('portainer.endpoints');
        }
        if (onAnalytics) {
          onAnalytics('kaas-agent');
        }
        setinitialName('');
      },
    });
  }

  return (
    <>
      <Formik<KaasCreateFormValues>
        initialValues={initialValues || { ...defaultValues, type: initialType }}
        onSubmit={(values, { resetForm }) =>
          onSubmit(values).then(() => {
            resetForm();
            return null;
          })
        }
        validationSchema={() => validationSchema(environmentNames)}
        validateOnMount
        enableReinitialize
      >
        <InnerKaasForm
          showTitle={showTitle}
          availableProviders={availableProviders}
          setApiKeyField={(providerName: string) =>
            setApiKeyField(providerName)
          }
          setInitialName={(name: string) => setinitialName(name)}
          setInitialType={(type: KaasProvider) => setInitialType(type)}
        />
      </Formik>
      {apiKeyField && (
        <CloudSettingsForm
          showCivo={apiKeyField === KaasProvider.CIVO}
          showLinode={apiKeyField === KaasProvider.LINODE}
          showDigitalOcean={apiKeyField === KaasProvider.DIGITAL_OCEAN}
          reroute={false}
        />
      )}
    </>
  );
}

// including the InnerKaasForm seems more complicated, but makes advanced formik work simple
// useFormikContext can handle form changes, conditionally render inputs, and manually set values
// useFormikContext can only work inside the Formik component
function InnerKaasForm({
  showTitle,
  availableProviders,
  setApiKeyField,
  setInitialName,
  setInitialType,
}: {
  showTitle?: boolean;
  availableProviders: KaasProvider[];
  setApiKeyField: (providerName: string) => void;
  setInitialName: (name: string) => void;
  setInitialType: (type: KaasProvider) => void;
}) {
  const { values, setFieldValue, errors, handleSubmit, isSubmitting, isValid } =
    useFormikContext<KaasCreateFormValues>();
  const { type: provider, region, name } = values;

  // cloudOptionsQuery updates then the provider updates
  const cloudOptionsQuery = useCloudProviderOptions(
    provider,
    availableProviders.includes(provider)
  );

  const [apiAvailable, setApiAvailable] = useState(
    availableProviders.includes(provider)
  );

  // if the available api keys or provider change, update apiAvailable
  useEffect(() => {
    if (availableProviders.includes(provider)) {
      setApiAvailable(true);
      setApiKeyField('');
    } else {
      setApiAvailable(false);
      setApiKeyField(provider);
    }
    setInitialType(provider);
  }, [availableProviders, provider]);

  // when the options change, set the field values to available options
  useEffect(() => {
    if (cloudOptionsQuery.data) {
      const newRegion = cloudOptionsQuery.data.regions[0].value;
      setFieldValue('region', newRegion);
      setFieldValue('nodeSize', cloudOptionsQuery.data.nodeSizes[0].value);
      setFieldValue(
        'kubernetesVersion',
        cloudOptionsQuery.data.kubernetesVersions[0]
      );
      if (cloudOptionsQuery.data?.networks) {
        const filteredNetworkOptions =
          cloudOptionsQuery.data.networks.get(newRegion) || [];
        if (filteredNetworkOptions.length > 0) {
          setFieldValue('networkId', filteredNetworkOptions[0].value);
        }
      }
    }
  }, [cloudOptionsQuery.data]);

  // handle when a user clicks another provider
  function onProviderChange(provider: KaasProvider) {
    setFieldValue('type', provider);
    if (availableProviders.includes(provider)) {
      setApiAvailable(true);
      setApiKeyField('');
    } else {
      setApiAvailable(false);
      setApiKeyField(provider);
    }
  }

  // when the name changes, update the initial name
  useEffect(() => {
    if (name) {
      setInitialName(name);
    }
  }, [name]);

  // when the region changes, update the selected network
  useEffect(() => {
    if (cloudOptionsQuery.data?.networks && region) {
      const filteredNetworkOptions =
        cloudOptionsQuery.data.networks.get(region) || [];
      if (filteredNetworkOptions.length > 0) {
        setFieldValue('networkId', filteredNetworkOptions[0].value);
      }
    }
  }, [
    region,
    cloudOptionsQuery.data?.networks,
    cloudOptionsQuery.data?.regions,
  ]);

  return (
    <Form className="form-horizontal" onSubmit={handleSubmit} noValidate>
      {showTitle && <FormSectionTitle>Environment details</FormSectionTitle>}
      <FormControl
        label="Name"
        tooltip="Name of cluster and environement"
        inputId="kaas-name"
        errors={errors.name}
      >
        <Field
          name="name"
          as={Input}
          id="kaas-name"
          data-cy="kaasCreateForm-nameInput"
          placeholder="e.g. MyClusterName"
        />
      </FormControl>
      <FormSectionTitle>Provision details</FormSectionTitle>
      <div className="form-group">
        <div className="col-sm-12">
          <Field
            name="type"
            as={BoxSelector}
            radioName="kaas-type"
            data-cy="kaasCreateForm-providerSelect"
            value={values.type}
            options={boxSelectorOptions}
            onChange={(provider: KaasProvider) => {
              onProviderChange(provider);
            }}
          />
        </div>
      </div>
      {/* loading */}
      {cloudOptionsQuery.isLoading && <Loading />}
      {/* helper message if there's no api key */}
      {!apiAvailable && (
        <div className="small text-warning" style={{ paddingBottom: '10px' }}>
          <i
            className="fa fa-exclamation-triangle orange-icon"
            aria-hidden="true"
            style={{ marginRight: '5px' }}
          />
          {`No API key found for ${providerTitles[provider]}. Save your ${providerTitles[provider]} API key below, or in the `}
          <Link to="portainer.settings.cloud" title="cloud settings">
            cloud settings
          </Link>
          .
        </div>
      )}
      {/* helper message if the api key is invalid */}
      {cloudOptionsQuery.isError && apiAvailable && (
        <div className="small text-warning" style={{ paddingBottom: '10px' }}>
          <i
            className="fa fa-exclamation-triangle orange-icon"
            aria-hidden="true"
            style={{ marginRight: '5px' }}
          />
          {`Error getting ${providerTitles[provider]} info. Go to `}
          <Link to="portainer.settings.cloud" title="cloud settings">
            cloud settings
          </Link>
          &nbsp;to ensure the API key is valid.
        </div>
      )}
      {/* create cluster fields */}
      {apiAvailable && cloudOptionsQuery.data && (
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
              id="kaa-nodeSize"
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
              min="1"
              id="kaas-nodeCount"
              placeholder="3"
            />
          </FormControl>
          {region && cloudOptionsQuery.data?.networks?.get(region) && (
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
                options={cloudOptionsQuery.data?.networks.get(region)}
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
              options={cloudOptionsQuery.data?.kubernetesVersions.map(
                (version) => ({
                  label: version,
                  value: version,
                })
              )}
            />
          </FormControl>
          <FormSectionTitle>Actions</FormSectionTitle>
          <div className="form-group">
            <div className="col-sm-12">
              <LoadingButton
                disabled={!isValid}
                isLoading={isSubmitting}
                loadingText="Provision in progress..."
              >
                <i className="fa fa-plus space-right" aria-hidden="true" />
                Provision environment
              </LoadingButton>
            </div>
          </div>
        </>
      )}
    </Form>
  );
}

function sendKaasProvisionAnalytics(values: KaasCreateFormValues) {
  trackEvent('portainer-endpoint-creation', {
    category: 'portainer',
    metadata: { type: 'agent', platform: 'kubernetes' },
  });
  trackEvent('provision-kaas-cluster', {
    category: 'kubernetes',
    metadata: {
      provider: values.type,
      region: values.region,
      'node-size': values.nodeSize,
      'node-count': values.nodeCount,
    },
  });
}

function getAvailableProviders(CloudApiKeys?: Partial<CloudApiKeys>) {
  const providers: KaasProvider[] = [];
  if (CloudApiKeys) {
    if (CloudApiKeys.CivoApiKey) {
      providers.push(KaasProvider.CIVO);
    }
    if (CloudApiKeys.LinodeToken) {
      providers.push(KaasProvider.LINODE);
    }
    if (CloudApiKeys.DigitalOceanToken) {
      providers.push(KaasProvider.DIGITAL_OCEAN);
    }
    return providers;
  }
  return providers;
}

export const KaasCreateFormAngular = react2angular(KaasCreateForm, [
  'view',
  'showTitle',
  'onUpdate',
  'onAnalytics',
]);
