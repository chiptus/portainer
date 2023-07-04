import { Loader2 } from 'lucide-react';
import { Formik } from 'formik';
import { useRouter } from '@uirouter/react';
import { useMemo } from 'react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { useMicroK8sOptions } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/queries';
import { queryClient } from '@/react-tools/react-query';

import { TextTip } from '@@/Tip/TextTip';
import { Card } from '@@/Card';

import {
  useAddonsQuery,
  useUpdateAddonsMutation,
} from '../../microk8s/addons.service';

import { K8sAddOnsForm } from './types';
import { AddonsForm } from './AddonsForm';

type Props = {
  currentVersion: string;
};

export function Addons({ currentVersion }: Props) {
  const router = useRouter();

  const environmentId = useEnvironmentId();
  const { data: environment, ...environmentQuery } =
    useEnvironment(environmentId);
  const { data: addonInfo, ...addonsQuery } = useAddonsQuery(
    environment?.Id,
    environment?.Status
  );
  const { data: microk8sOptions, ...microk8sOptionsQuery } =
    useMicroK8sOptions();
  const addonsUpdateMutation = useUpdateAddonsMutation();

  const addonOptions: string[] = useMemo(() => {
    const addons: string[] = [];
    microk8sOptions?.availableAddons.forEach((a) => {
      const kubeVersion = parseFloat(currentVersion.split('/')[0]);
      const versionAvailableFrom = parseFloat(a.versionAvailableFrom);
      if (kubeVersion >= versionAvailableFrom) {
        addons.push(a.label);
      }
    });
    return addons;
  }, [microk8sOptions?.availableAddons, currentVersion]);

  // the optional addons (excluding the required ones)
  const initialValues = {
    addons:
      addonInfo?.addons
        .filter((addon) => addon.status === 'enabled')
        .filter((addon) => addonOptions.includes(addon.name)) // show only installable addons
        .map((addon) => addon.name) || [],
  };

  if (addonsQuery.isError || environmentQuery.isError) {
    return <TextTip color="orange">Unable to get addons</TextTip>;
  }

  if (microk8sOptionsQuery.isError) {
    return <TextTip color="orange">Unable to get addon options.</TextTip>;
  }

  return (
    <Card>
      {addonsQuery.isLoading && (
        <div className="vertical-center text-muted">
          <Loader2 className="h-4 animate-spin-slow" />
          Loading addons...
        </div>
      )}
      {addonsQuery.isSuccess && (
        <Formik
          initialValues={initialValues}
          onSubmit={(values) => handleUpdateAddons(values)}
          validateOnMount
          enableReinitialize
        >
          {AddonsForm}
        </Formik>
      )}
    </Card>
  );

  function handleUpdateAddons(values: K8sAddOnsForm) {
    return new Promise((resolve) => {
      const payload = {
        addons: values.addons.map((addon) => addon),
      };
      addonsUpdateMutation.mutate(
        {
          environmentID: environmentId,
          credentialID: environment?.CloudProvider.CredentialID || 0,
          payload,
        },
        {
          onSuccess: () => {
            resolve([true, 0]);
            notifySuccess('Success', 'Addons update requested successfully');
            queryClient.refetchQueries(['environments', environmentId]);
            router.stateService.reload();
          },
          onError: (error) => {
            resolve([false, 0]);
            notifyError('Error requesting addons update', error as Error);
          },
        }
      );
    });
  }
}
