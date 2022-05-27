import { useState } from 'react';
import { Formik } from 'formik';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import { useSettings } from '@/portainer/settings/queries';
import { KaasProvider, Credential } from '@/portainer/settings/cloud/types';
import { AnalyticsStateKey } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/types';
import { Environment } from '@/portainer/environments/types';

import { useCreateKaasCluster } from '../queries';
import { CreateGKEClusterFormValues, GKEKaasInfo } from '../types';
import { sendKaasProvisionAnalytics } from '../utils';

import { GKECreateClusterForm } from './GKECreateClusterForm';
import { validationSchema } from './validation';

type Props = {
  name: string;
  setName: (name: string) => void;
  provider: KaasProvider;
  credentials: Credential[];
  onUpdate?: () => void;
  onCreate?(environment: Environment, analytics: AnalyticsStateKey): void;
};

export function GKECreateClusterFormContainer({
  name,
  setName,
  provider,
  credentials,
  onUpdate,
  onCreate,
}: Props) {
  const [gkeKaasInfo, setGKEKaasInfo] = useState<GKEKaasInfo>();
  const [vCPUCount, setvCPUCount] = useState<number>(2);

  const { state } = useCurrentStateAndParams();
  const router = useRouter();

  const settingsQuery = useSettings();
  const createKaasCluster = useCreateKaasCluster();

  const initialValues = {
    nodeCount: 3,
    kubernetesVersion: '',
    nodeSize: '',
    cpu: 2,
    ram: 4,
    hdd: 100,
    region: '',
    networkId: '',
    credentialId: credentials[0].id,
    meta: {
      groupId: 1,
      tagIds: [],
    },
  };

  return (
    <Formik<CreateGKEClusterFormValues>
      validateOnChange
      initialValues={initialValues}
      onSubmit={(values) => onSubmit(values)}
      validationSchema={() => validationSchema(vCPUCount, gkeKaasInfo)}
      validateOnMount
    >
      <GKECreateClusterForm
        credentials={credentials}
        provider={provider}
        name={name}
        setGKEKaasInfo={setGKEKaasInfo}
        setvCPUCount={setvCPUCount}
      />
    </Formik>
  );

  async function onSubmit(formValues: CreateGKEClusterFormValues) {
    if (settingsQuery.data?.EnableTelemetry) {
      sendKaasProvisionAnalytics(formValues, provider);
    }
    let payload;
    if (formValues.nodeSize === 'custom') {
      payload = {
        ...formValues,
        // credentialId is sometimes converted to a string
        credentialId: Number(formValues.credentialId),
        name,
      };
    } else {
      // omit the cpu and ram values
      payload = {
        region: formValues.region,
        networkId: formValues.networkId,
        nodeSize: formValues.nodeSize,
        hdd: formValues.hdd,
        kubernetesVersion: formValues.kubernetesVersion,
        nodeCount: formValues.nodeCount,
        credentialId: Number(formValues.credentialId),
        meta: formValues.meta,
        name,
      };
    }
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
