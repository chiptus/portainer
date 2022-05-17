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
import { useCloudCredentials } from '@/portainer/settings/cloud/cloudSettings.service';
import { KaasProvider, Credential } from '@/portainer/settings/cloud/types';
import { useEnvironmentList } from '@/portainer/environments/queries/useEnvironment';

import { useCloudProviderOptions, useCreateKaasCluster } from './queries';
import {
  KaasCreateFormInitialValues,
  KaasCreateFormValues,
  CredentialProviderInfo,
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
  googlecloud: 'Google Cloud',
  aws: 'AWS',
  azure: 'Azure',
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
  const cloudCredentialsQuery = useCloudCredentials();
  const [availableProviders, setAvailableProviders] = useState<KaasProvider[]>(
    []
  );
  const cloudOptionsQuery = useCloudProviderOptions(
    cloudCredentialsQuery?.data?.[0]
  );

  const environmentsQuery = useEnvironmentList();
  const environmentNames = environmentsQuery.environments?.map(
    (env) => env.Name
  );
  const createKaasCluster = useCreateKaasCluster();

  const [initialValues, setInitialValues] = useState<
    KaasCreateFormValues | undefined
  >(undefined);
  // remember some form values
  const [initialName, setinitialName] = useState<string>(
    KaasCreateFormInitialValues.name
  );
  const initialType = availableProviders[0] || KaasProvider.CIVO;

  // when the api keys change, update the available providers
  useEffect(() => {
    if (cloudCredentialsQuery?.data?.length) {
      setAvailableProviders(getAvailableProviders(cloudCredentialsQuery?.data));
    }
  }, [cloudCredentialsQuery?.data]);

  // set the initial form values when they are available
  useEffect(() => {
    if (cloudOptionsQuery.data && availableProviders[0]) {
      // only set the initial values once
      if (!initialValues && cloudOptionsQuery.isSuccess) {
        const defaultRegion = cloudOptionsQuery.data?.regions[0].value;
        const credential = cloudCredentialsQuery?.data?.find(
          (credential) => credential.provider === initialType
        );
        setInitialValues({
          name: initialName,
          nodeCount: KaasCreateFormInitialValues.nodeCount,
          type: initialType,
          kubernetesVersion: cloudOptionsQuery.data?.kubernetesVersions[0],
          nodeSize: cloudOptionsQuery.data?.nodeSizes[0].value,
          region: defaultRegion,
          networkId: cloudOptionsQuery.data?.networks?.get(defaultRegion)?.at(0)
            ?.value,
          credentialId: credential?.id,
        });
      }
    }
  }, [
    cloudOptionsQuery.data,
    cloudOptionsQuery.isSuccess,
    initialName,
    initialType,
    initialValues,
    availableProviders,
    cloudCredentialsQuery.data,
  ]);

  // handle the submit with current form values
  async function onSubmit(formValues: KaasCreateFormValues) {
    if (settingsQuery.data?.EnableTelemetry) {
      sendKaasProvisionAnalytics(formValues);
    }
    const formValuesToSubmit = {
      ...formValues,
      credentialId: Number(formValues.credentialId),
    };
    createKaasCluster.mutate(formValuesToSubmit, {
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
        setInitialName={(name: string) => setinitialName(name)}
        credentials={cloudCredentialsQuery.data}
      />
    </Formik>
  );
}

// including the InnerKaasForm seems more complicated, but makes advanced formik work simple
// useFormikContext can handle form changes, conditionally render inputs, and manually set values
// useFormikContext can only work inside the Formik component
function InnerKaasForm({
  showTitle,
  availableProviders,
  setInitialName,
  credentials,
}: {
  showTitle?: boolean;
  availableProviders: KaasProvider[];
  setInitialName: (name: string) => void;
  credentials?: Credential[];
}) {
  const { values, setFieldValue, errors, handleSubmit, isSubmitting, isValid } =
    useFormikContext<KaasCreateFormValues>();
  const { type: provider, region, name, credentialId } = values;

  const [defaultCredential, setDefaultCredential] = useState<Credential>();
  const [providerAvailable, setProviderAvailable] = useState<boolean>(false);

  useEffect(() => {
    const credential = credentials?.find(
      (credential) => credential.provider === provider
    );
    setDefaultCredential(credential);
    setProviderAvailable(!!credential);
    setFieldValue('credentialId', credential?.id);
  }, [provider, credentials, setFieldValue]);

  // cloudOptionsQuery updates then the provider updates
  const cloudOptionsQuery = useCloudProviderOptions(defaultCredential);
  const [credsDropdown, setCredsDropdown] = useState<CredentialProviderInfo>(
    new Map()
  );

  useEffect(() => {
    const credential = credentials?.find(
      (credential) => credential.id === Number(credentialId)
    );
    setDefaultCredential(credential);
  }, [credentialId, credentials]);

  useEffect(() => {
    const creds: CredentialProviderInfo = new Map();
    credentials?.forEach((credential) => {
      const providerCreds = creds.get(credential.provider) || [];
      providerCreds.push({ value: credential.id, label: credential.name });
      creds.set(credential.provider, providerCreds);
    });
    setCredsDropdown(creds);
  }, [credentials]);

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
  }, [cloudOptionsQuery.data, setFieldValue]);

  // handle when a user clicks another provider
  function onProviderChange(provider: KaasProvider) {
    setFieldValue('type', provider);
    if (availableProviders.includes(provider)) {
      const credential = credentials?.find(
        (credential) => credential.provider === provider
      );
      setDefaultCredential(credential);
      setProviderAvailable(true);
    } else {
      setDefaultCredential(undefined);
      setProviderAvailable(false);
    }
  }

  // when the name changes, update the initial name
  useEffect(() => {
    if (name) {
      setInitialName(name);
    }
  }, [name, setInitialName]);

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
    setFieldValue,
  ]);

  return (
    <Form className="form-horizontal" onSubmit={handleSubmit} noValidate>
      {showTitle && <FormSectionTitle>Environment details</FormSectionTitle>}
      <FormControl
        label="Name"
        tooltip="Name of cluster and environment"
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
      {!providerAvailable && (
        <div className="small text-warning" style={{ paddingBottom: '10px' }}>
          <i
            className="fa fa-exclamation-triangle orange-icon"
            aria-hidden="true"
            style={{ marginRight: '5px' }}
          />
          {`No API key found for ${providerTitles[provider]}. Save your ${providerTitles[provider]} API key in the `}
          <Link to="portainer.settings.cloud" title="cloud settings">
            cloud settings
          </Link>
          .
        </div>
      )}
      {/* helper message if the api key is invalid */}
      {cloudOptionsQuery.isError && providerAvailable && (
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
      {providerAvailable && cloudOptionsQuery.data && (
        <>
          <FormControl
            label="Choose Credentials"
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
              options={credsDropdown?.get(provider) || []}
            />
          </FormControl>
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

function getAvailableProviders(credentials: Credential[]) {
  const providers: KaasProvider[] = [];
  credentials.forEach((credential) => {
    if (providers.indexOf(credential.provider) === -1) {
      providers.push(credential.provider);
    }
  });
  return providers;
}

export const KaasCreateFormAngular = react2angular(KaasCreateForm, [
  'view',
  'showTitle',
  'onUpdate',
  'onAnalytics',
]);
