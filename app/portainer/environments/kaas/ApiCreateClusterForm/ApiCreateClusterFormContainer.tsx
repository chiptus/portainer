import { Formik } from 'formik';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import { useSettings } from '@/portainer/settings/queries';
import { KaasProvider, Credential } from '@/portainer/settings/cloud/types';
import { AnalyticsStateKey } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/types';
import { Environment } from '@/portainer/environments/types';

import { useCreateKaasCluster } from '../queries';
import { CreateApiClusterFormValues } from '../types';
import { sendKaasProvisionAnalytics } from '../utils';

import { validationSchema } from './validation';
import { ApiCreateClusterForm } from './ApiCreateClusterForm';

type Props = {
  name: string;
  setName: (name: string) => void;
  provider: KaasProvider;
  credentials: Credential[];
  onUpdate?: () => void;
  onCreate?(environment: Environment, analytics: AnalyticsStateKey): void;
};

export function ApiCreateClusterFormContainer({
  name,
  setName,
  provider,
  credentials,
  onUpdate,
  onCreate,
}: Props) {
  const { state } = useCurrentStateAndParams();
  const router = useRouter();

  const settingsQuery = useSettings();
  const createKaasCluster = useCreateKaasCluster();

  const initialValues = {
    nodeCount: 3,
    kubernetesVersion: '',
    nodeSize: '',
    region: '',
    networkId: '',
    credentialId: credentials[0].id,
    meta: {
      groupId: 1,
      tagIds: [],
    },
  };

  return (
    <Formik<CreateApiClusterFormValues>
      initialValues={initialValues}
      onSubmit={(values, { resetForm }) =>
        onSubmit(values).then(() => {
          resetForm();
          return null;
        })
      }
      validationSchema={() => validationSchema()}
      validateOnMount
      enableReinitialize
    >
      <ApiCreateClusterForm
        credentials={credentials}
        provider={provider}
        name={name}
      />
    </Formik>
  );

  async function onSubmit(formValues: CreateApiClusterFormValues) {
    if (settingsQuery.data?.EnableTelemetry) {
      sendKaasProvisionAnalytics(formValues, provider);
    }
    const payload = {
      ...formValues,
      // credentialId is sometimes converted to a string
      credentialId: Number(formValues.credentialId),
      name,
    };
    createKaasCluster.mutate(
      { payload, provider },
      {
        onSuccess: (environment) => {
          if (onUpdate) {
            onUpdate();
          }
          if (state.name === 'portainer.endpoints.new') {
            router.stateService.go('portainer.endpoints');
          } else {
            setName('');
          }
          if (onCreate) {
            onCreate(environment, 'kaasAgent');
          }
        },
      }
    );
  }
}
