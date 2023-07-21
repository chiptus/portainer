import { Loader2 } from 'lucide-react';
import { Formik } from 'formik';
import { useRouter } from '@uirouter/react';
import { useMemo } from 'react';
import { SchemaOf, object, string } from 'yup';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { useMicroK8sOptions } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/queries';
import { queryClient } from '@/react-tools/react-query';
import { AddonOptionInfo } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/types';

import { TextTip } from '@@/Tip/TextTip';
import { Card } from '@@/Card';
import { confirmUpdate } from '@@/modals/confirm';
import { Icon } from '@@/Icon';

import { K8sAddOnsForm } from './types';
import { useAddonsQuery, useUpdateAddonsMutation } from './addons.service';
import { AddonsForm } from './AddonsForm';
import { addonsValidation } from './addonsValidation';

export function addonsFormValidation(
  addonOptionsInfo: AddonOptionInfo[],
  initialValues?: K8sAddOnsForm
): SchemaOf<K8sAddOnsForm> {
  return object({
    addons: addonsValidation(addonOptionsInfo, initialValues),
    currentVersion: string().required(),
  });
}

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
  const addonOptions: AddonOptionInfo[] = useMemo(() => {
    const addons: AddonOptionInfo[] = [];
    microk8sOptions?.availableAddons.forEach((a) => {
      const kubeVersion = parseFloat(currentVersion.split('/')[0]);
      const versionAvailableFrom = parseFloat(a.versionAvailableFrom);
      const versionAvailableTo = parseFloat(a.versionAvailableTo);
      if (
        kubeVersion >= versionAvailableFrom &&
        (Number.isNaN(versionAvailableTo) || kubeVersion <= versionAvailableTo)
      ) {
        addons.push(a);
      }
    });
    return addons;
  }, [microk8sOptions?.availableAddons, currentVersion]);

  // the optional addons (excluding the required ones)
  const initialValues: K8sAddOnsForm = {
    addons:
      addonInfo?.addons
        .filter((addonInfo) => addonInfo.status === 'enabled')
        .filter((addonInfo) =>
          addonOptions.find(
            (addonOption) => addonOption.label === addonInfo.name
          )
        ) // show only installable addons (not the required ones)
        .map((addonInfo) => ({
          name: addonInfo.name,
          arguments: addonInfo.arguments,
          repository: addonInfo.repository,
        })) || // format to form values
      [],
    currentVersion,
  };

  if (addonsQuery.isLoading || microk8sOptionsQuery.isLoading) {
    return (
      <Card>
        <div className="vertical-center text-muted text-sm">
          <Icon icon={Loader2} className="animate-spin-slow" />
          Loading addons...
        </div>
      </Card>
    );
  }

  if (microk8sOptionsQuery.isError || addonsQuery.isError) {
    return (
      <Card>
        {microk8sOptionsQuery.isError && (
          <TextTip color="red">Unable to get addon options</TextTip>
        )}
        {(addonsQuery.isError || environmentQuery.isError) && (
          <TextTip color="red">Unable to get addons</TextTip>
        )}
      </Card>
    );
  }

  return (
    <Card>
      {microk8sOptions && (
        <Formik
          initialValues={initialValues}
          onSubmit={handleUpdateAddons}
          validationSchema={addonsFormValidation(
            microk8sOptions.availableAddons,
            initialValues
          )}
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
          addonsUpdateMutation.mutate(
            {
              environmentID: environmentId,
              credentialID: environment?.CloudProvider?.CredentialID || 0,
              payload: { addons: values.addons },
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