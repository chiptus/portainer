import { Formik } from 'formik';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import { useSettings } from '@/portainer/settings/queries';
import { KaasProvider, Credential } from '@/portainer/settings/cloud/types';

import { useCreateKaasCluster } from '../queries';
import { CreateAzureClusterFormValues } from '../types';
import { sendKaasProvisionAnalytics } from '../utils';

import { validationSchema } from './validation';
import { AzureCreateClusterForm } from './AzureCreateClusterForm';

type Props = {
  name: string;
  setName: (name: string) => void;
  provider: KaasProvider;
  credentials: Credential[];
  onUpdate?: () => void;
  onAnalytics?: (eventName: string) => void;
};

export function AzureCreateClusterFormContainer({
  name,
  setName,
  provider,
  credentials,
  onUpdate,
  onAnalytics,
}: Props) {
  const { state } = useCurrentStateAndParams();
  const router = useRouter();

  const settingsQuery = useSettings();
  const createKaasCluster = useCreateKaasCluster();

  const initialValues = {
    resourceGroup: '',
    resourceGroupName: '',
    tier: 'Free',
    poolName: '',
    nodeCount: 3,
    kubernetesVersion: '',
    nodeSize: '',
    region: '',
    credentialId: credentials[0].id,
    dnsPrefix: '',
    availabilityZones: [],
    resourceGroupInput: 'select',
    meta: {
      groupId: 1,
      tagIds: [],
    },
  };

  return (
    <Formik<CreateAzureClusterFormValues>
      initialValues={initialValues}
      onSubmit={(values) => onSubmit(values)}
      validationSchema={() => validationSchema()}
      validateOnMount
      enableReinitialize
    >
      <AzureCreateClusterForm
        credentials={credentials}
        provider={provider}
        name={name}
      />
    </Formik>
  );

  async function onSubmit(formValues: CreateAzureClusterFormValues) {
    if (settingsQuery.data?.EnableTelemetry) {
      sendKaasProvisionAnalytics(formValues, provider);
    }
    let resourceGroup;
    let resourceGroupName;
    if (formValues.resourceGroupInput === 'select') {
      resourceGroup = formValues.resourceGroup;
      resourceGroupName = '';
    } else {
      resourceGroup = '';
      resourceGroupName = formValues.resourceGroupName;
    }
    const payload = {
      ...formValues,
      // credentialId is sometimes converted to a string
      credentialId: Number(formValues.credentialId),
      resourceGroup,
      resourceGroupName,
      name,
    };
    createKaasCluster.mutate(
      { payload, provider },
      {
        onSuccess: () => {
          if (onUpdate) {
            onUpdate();
          }
          if (state.name === 'portainer.endpoints.new') {
            router.stateService.go('portainer.endpoints');
          } else {
            setName('');
          }
          if (onAnalytics) {
            onAnalytics('kaas-agent');
          }
        },
      }
    );
  }
}
