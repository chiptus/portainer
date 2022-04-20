import { v4 as uuid } from 'uuid';
import clsx from 'clsx';
import { Formik } from 'formik';
import { useState } from 'react';
import { useMutation } from 'react-query';

import { BoxSelectorItem } from '@/portainer/components/BoxSelector/BoxSelectorItem';
import { r2a } from '@/react-tools/react2angular';
import { createRemoteEndpoint } from '@/portainer/environments/environment.service/create';
import { EnvironmentCreationTypes } from '@/portainer/environments/types';
import { error as notifyError } from '@/portainer/services/notifications';
import { baseHref } from '@/portainer/helpers/pathHelper';

import styles from './WizardNomad.module.css';
import { validationSchema } from './validation';
import { EdgeInfo, FormValues } from './types';
import { NomadForm } from './NomadForm';

interface Props {
  onUpdate(): void;
  agentVersion: string;
}

export function WizardNomad({ onUpdate, agentVersion }: Props) {
  const initialValues: FormValues = {
    name: '',
    token: '',
    portainerUrl: defaultPortainerUrl(),
    pollFrequency: 0,
    allowSelfSignedCertificates: true,
    authEnabled: true,
    envVars: '',
  };

  const [edgeInfo, setEdgeInfo] = useState<EdgeInfo>();

  const createMutation = useMutation((values: FormValues) =>
    createRemoteEndpoint(
      values.name,
      EnvironmentCreationTypes.EdgeAgentEnvironment,
      { url: values.portainerUrl, checkinInterval: values.pollFrequency }
    )
  );

  return (
    <>
      <div className={clsx(styles.options)}>
        <BoxSelectorItem
          option={{
            description: 'Portainer Edge Agent',
            icon: 'fa fa-cloud',
            id: 'id',
            label: 'Edge Agent',
            value: 'edge',
          }}
          onChange={() => {}}
          radioName="radio"
          selectedValue="edge"
        />
      </div>

      <Formik<FormValues>
        initialValues={initialValues}
        validationSchema={() => validationSchema()}
        onSubmit={handleSubmit}
        onReset={handleReset}
        validateOnMount
      >
        <NomadForm
          agentVersion={agentVersion}
          isSubmitting={createMutation.isLoading}
          edgeInfo={edgeInfo}
        />
      </Formik>
    </>
  );

  function handleReset() {
    setEdgeInfo(undefined);
  }

  function handleSubmit(values: FormValues) {
    createMutation.mutate(values, {
      onError(error) {
        notifyError('Failure', error as Error, 'Unable to create endpoint');
      },
      onSuccess(environment) {
        setEdgeInfo({ key: environment.EdgeKey, id: uuid() });
        onUpdate();
      },
    });
  }
}

function defaultPortainerUrl() {
  const baseHREF = baseHref();
  return window.location.origin + (baseHREF !== '/' ? baseHREF : '');
}

export const WizardNomadAngular = r2a(WizardNomad, [
  'onUpdate',
  'agentVersion',
]);
