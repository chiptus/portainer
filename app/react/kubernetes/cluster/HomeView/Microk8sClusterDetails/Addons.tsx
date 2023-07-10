import { Loader2 } from 'lucide-react';
import { Formik } from 'formik';
import { useRouter } from '@uirouter/react';
import { useMemo } from 'react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { useMicroK8sOptions } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/queries';
import { queryClient } from '@/react-tools/react-query';
import { AddOnOption } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/Microk8sCreateClusterForm/AddonSelector';

import { TextTip } from '@@/Tip/TextTip';
import { Card } from '@@/Card';
import { confirmUpdate } from '@@/modals/confirm';
import { Icon } from '@@/Icon';

import {
  useAddonsQuery,
  useUpdateAddonsMutation,
} from '../../microk8s/addons.service';

import { K8sAddOnsForm } from './types';
import { AddonsForm } from './AddonsForm';

export function Addons() {
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

  const currentVersion = addonInfo?.currentVersion ?? '';
  const addonOptions: AddOnOption[] = useMemo(() => {
    const addons: AddOnOption[] = [];
    microk8sOptions?.availableAddons.forEach((a) => {
      const kubeVersion = parseFloat(currentVersion.split('/')[0]);
      const versionAvailableFrom = parseFloat(a.versionAvailableFrom);
      if (kubeVersion >= versionAvailableFrom) {
        addons.push({ name: a.label, type: a.type });
      }
    });
    return addons;
  }, [microk8sOptions?.availableAddons, currentVersion]);

  // the optional addons (excluding the required ones)
  const initialValues = {
    addons:
      addonInfo?.addons
        .filter((addon) => addon.status === 'enabled')
        .filter((addon) => addonOptions.find((a) => a.name === addon.name)) // show only installable addons
        .map((addon) => ({ name: addon.name, type: addon.repository })) || [],
    currentVersion,
  };

  return (
    <Card>
      {microk8sOptionsQuery.isError && (
        <TextTip color="red">Unable to get addon options</TextTip>
      )}
      {(addonsQuery.isError || environmentQuery.isError) && (
        <TextTip color="red">Unable to get addons</TextTip>
      )}
      {addonsQuery.isLoading && (
        <div className="vertical-center text-muted text-sm">
          <Icon icon={Loader2} className="animate-spin-slow" />
          Loading addons...
        </div>
      )}
      {addonsQuery.isSuccess && (
        <Formik
          initialValues={initialValues}
          onSubmit={handleUpdateAddons}
          validateOnMount
          enableReinitialize
        >
          {AddonsForm}
        </Formik>
      )}
    </Card>
  );

  async function handleUpdateAddons(values: K8sAddOnsForm) {
    confirmUpdate(
      'Are you sure you want to apply changes to addons?',
      (confirmed) => {
        if (confirmed) {
          const payload = {
            addons: values.addons.map((addon) => addon.name),
          };
          addonsUpdateMutation.mutate(
            {
              environmentID: environmentId,
              credentialID: environment?.CloudProvider.CredentialID || 0,
              payload,
            },
            {
              onSuccess: () => {
                notifySuccess(
                  'Success',
                  'Request to update addons successfully submitted'
                );
                queryClient.refetchQueries(['environments', environmentId]);
                router.stateService.reload();
              },
              onError: (error) => {
                notifyError('Error requesting addons update', error as Error);
              },
            }
          );
        }
      }
    );
  }
}
