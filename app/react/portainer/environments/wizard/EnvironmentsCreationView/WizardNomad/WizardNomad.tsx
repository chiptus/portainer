import { v4 as uuid } from 'uuid';
import { Formik } from 'formik';
import { useState } from 'react';

import { BoxSelectorItem } from '@/portainer/components/BoxSelector/BoxSelectorItem';
import { error as notifyError } from '@/portainer/services/notifications';
import { baseHref } from '@/portainer/helpers/pathHelper';
import { useCreateEdgeAgentEnvironmentMutation } from '@/portainer/environments/queries/useCreateEnvironmentMutation';
import { Environment } from '@/portainer/environments/types';

import { AnalyticsStateKey } from '../types';

import { validationSchema } from './validation';
import { EdgeInfo, FormValues } from './types';
import { NomadForm } from './NomadForm';

interface Props {
  onCreate(environment: Environment, analytics: AnalyticsStateKey): void;
}

export function WizardNomad({ onCreate }: Props) {
  const initialValues: FormValues = {
    name: '',
    token: '',
    portainerUrl: defaultPortainerUrl(),
    pollFrequency: 0,
    allowSelfSignedCertificates: true,
    authEnabled: true,
    envVars: '',
    useAsyncMode: false,
    meta: {
      groupId: 1,
      tagIds: [],
    },
  };

  const [edgeInfo, setEdgeInfo] = useState<EdgeInfo>();

  const createMutation = useCreateEdgeAgentEnvironmentMutation();

  return (
    <div className="form-horizontal">
      <div className="form-group">
        <div className="col-sm-12">
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
      </div>

      <Formik<FormValues>
        initialValues={initialValues}
        validationSchema={() => validationSchema()}
        onSubmit={handleSubmit}
        onReset={handleReset}
        validateOnMount
      >
        <NomadForm
          isSubmitting={createMutation.isLoading}
          edgeInfo={edgeInfo}
        />
      </Formik>
    </div>
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
        onCreate(environment, 'nomadEdgeAgent');
      },
    });
  }
}

function defaultPortainerUrl() {
  const baseHREF = baseHref();
  return window.location.origin + (baseHREF !== '/' ? baseHREF : '');
}
