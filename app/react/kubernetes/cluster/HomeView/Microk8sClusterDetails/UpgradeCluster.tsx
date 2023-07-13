import { Formik, Form, Field } from 'formik';
import { useRouter } from '@uirouter/react';
import { Loader2 } from 'lucide-react';

import { notifySuccess } from '@/portainer/services/notifications';
import { useAuthorizations } from '@/react/hooks/useUser';
import { useAnalytics } from '@/react/hooks/useAnalytics';
import { useCurrentEnvironment } from '@/react/hooks/useCurrentEnvironment';
import { K8sDistributionType } from '@/react/portainer/environments/types';

import { confirm } from '@@/modals/confirm';
import { Input } from '@@/form-components/Input';
import { Card } from '@@/Card';
import { FormControl } from '@@/form-components/FormControl';
import { LoadingButton } from '@@/buttons';
import { ModalType } from '@@/modals';
import { TextTip } from '@@/Tip/TextTip';
import { Icon } from '@@/Icon';

import {
  useAddonsQuery,
  useUpgradeClusterMutation,
} from '../../microk8s/addons/addons.service';

export type K8sUpgradeType = {
  kubeVersion: string;
};

export function UpgradeCluster() {
  const router = useRouter();

  const { data: environment, ...environmentQuery } = useCurrentEnvironment();
  const statusQuery = useAddonsQuery(environment?.Id, environment?.Status);
  const { currentVersion, kubernetesVersions } = statusQuery.data || {};
  const isProcessing =
    environment?.StatusMessage?.operationStatus === 'processing';

  const isAllowed = useAuthorizations(['K8sClusterW']);

  const initialValues: K8sUpgradeType = {
    kubeVersion: currentVersion || '',
  };

  const upgradeClusterMutation = useUpgradeClusterMutation();

  const index =
    kubernetesVersions?.findIndex((v) => v.value === currentVersion) || 0;
  const nextVersion = kubernetesVersions?.[index - 1]?.value || currentVersion;

  const { trackEvent } = useAnalytics();

  return (
    <Card>
      {statusQuery.isError && (
        <TextTip color="red">Unable to Kubernetes version</TextTip>
      )}
      {environmentQuery.isError && (
        <TextTip color="red">Unable to load environment</TextTip>
      )}
      {(statusQuery.isLoading || environmentQuery.isLoading) && (
        <div className="vertical-center text-muted text-sm">
          <Icon icon={Loader2} className="animate-spin-slow" />
          Loading Kubernetes version...
        </div>
      )}
      {statusQuery.isSuccess && environment && (
        <Formik
          initialValues={initialValues}
          onSubmit={handleUpgradeCluster}
          validateOnMount
          enableReinitialize
        >
          {({ isSubmitting }) => (
            <Form className="form-horizontal">
              <FormControl
                label="Kubernetes version"
                tooltip="Kubernetes version running on the cluster. Upgrades can only be performed to the next version number (where a newer version is available)."
                inputId="kubeVersion"
              >
                <Field
                  name="kubeVersion"
                  as={Input}
                  id="kubeVersion-input"
                  disabled
                />
              </FormControl>
              <LoadingButton
                isLoading={isSubmitting}
                loadingText="Upgrading..."
                type="submit"
                color="secondary"
                className="!ml-0"
                onClick={() => {}}
                disabled={
                  nextVersion === currentVersion ||
                  upgradeClusterMutation.isLoading ||
                  !isAllowed ||
                  isProcessing ||
                  statusQuery.isRefetching
                }
              >
                {nextVersion === currentVersion
                  ? 'No upgrade currently available'
                  : `Upgrade to ${nextVersion}`}
              </LoadingButton>
            </Form>
          )}
        </Formik>
      )}
    </Card>
  );

  async function handleUpgradeCluster() {
    if (!environment) return;
    const confirmed = await confirm({
      title: 'Are you sure?',
      message:
        'Are you sure you want to upgrade the cluster? This may cause the cluster to be unavailable during the upgrade process.',
      cancelButtonLabel: 'Cancel',
      modalType: ModalType.Warn,
    });
    if (confirmed) {
      upgradeClusterMutation.mutate(
        { environmentID: environment?.Id, nextVersion: nextVersion || '' },
        {
          onSuccess: () => {
            notifySuccess('Success', 'Cluster upgrade requested successfully');
            trackEvent('upgrade-k8s-cluster', {
              category: 'kubernetes',
              metadata: {
                provider: K8sDistributionType.MICROK8S,
                nextVersion,
              },
            });
            router.stateService.reload();
          },
        }
      );
    }
  }
}
