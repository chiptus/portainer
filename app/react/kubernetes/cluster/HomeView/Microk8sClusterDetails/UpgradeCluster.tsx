import { Loader2 } from 'lucide-react';
import { Formik, Form, Field } from 'formik';
import { useRouter } from '@uirouter/react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { notifySuccess } from '@/portainer/services/notifications';
import { useAuthorizations } from '@/react/hooks/useUser';
import { useAnalytics } from '@/react/hooks/useAnalytics';
import { K8sDistributionType } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/types';
import { useEnvironment } from '@/react/portainer/environments/queries';

import { confirm } from '@@/modals/confirm';
import { Input } from '@@/form-components/Input';
import { Card } from '@@/Card';
import { FormControl } from '@@/form-components/FormControl';
import { LoadingButton } from '@@/buttons';
import { ModalType } from '@@/modals';
import { Option } from '@@/form-components/Input/Select';

import { useUpgradeClusterMutation } from '../../microk8s/addons.service';

export type K8sUpgradeType = {
  kubeVersion: string;
};

type Props = {
  kubernetesVersions?: Option<string>[];
  currentVersion?: string;
  statusQuery: {
    isLoading: boolean;
    isError: boolean;
  };
};

export function UpgradeCluster({
  currentVersion,
  kubernetesVersions,
  statusQuery,
}: Props) {
  const router = useRouter();

  const environmentId = useEnvironmentId();
  const environmentQuery = useEnvironment(environmentId);
  const { data: environment } = environmentQuery;
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
    <Formik
      initialValues={initialValues}
      onSubmit={handleUpgradeCluster}
      validateOnMount
      enableReinitialize
    >
      <Form className="form-horizontal">
        <Card>
          {statusQuery.isError && 'Unable to get kubernetes status'}
          <FormControl
            label="Kubernetes version"
            tooltip="Kubernetes version running on the cluster. Upgrades can only be performed to the next version number (where a newer version is available)."
            inputId="kubeVersion"
          >
            {statusQuery.isLoading && (
              <div className="vertical-center text-muted pt-2">
                <Loader2 className="h-4 animate-spin-slow" />
                Loading version...
              </div>
            )}
            {!statusQuery.isLoading && (
              <Field
                name="kubeVersion"
                as={Input}
                id="kubeVersion-input"
                disabled
              />
            )}
          </FormControl>
          {!statusQuery.isLoading && (
            <LoadingButton
              isLoading={statusQuery.isLoading}
              loadingText="Upgrading..."
              type="submit"
              color="secondary"
              className="!ml-0"
              onClick={() => {}}
              disabled={
                nextVersion === currentVersion ||
                upgradeClusterMutation.isLoading ||
                !isAllowed ||
                isProcessing
              }
            >
              {nextVersion === currentVersion
                ? 'No upgrade currently available'
                : `Upgrade to ${nextVersion}`}
            </LoadingButton>
          )}
        </Card>
      </Form>
    </Formik>
  );

  async function handleUpgradeCluster() {
    const confirmed = await confirm({
      title: 'Are you sure?',
      message:
        'Are you sure you want to upgrade the cluster? This may cause the cluster to be unavailable during the upgrade process.',
      cancelButtonLabel: 'Cancel',
      modalType: ModalType.Warn,
    });
    if (confirmed) {
      upgradeClusterMutation.mutate(
        { environmentID: environmentId, nextVersion: nextVersion || '' },
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
