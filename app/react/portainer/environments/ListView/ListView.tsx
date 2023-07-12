import { useStore } from 'zustand';

import { environmentStore } from '@/react/hooks/current-environment-store';
import { useAnalytics } from '@/react/hooks/useAnalytics';

import { PageHeader } from '@@/PageHeader';

import { Environment } from '../types';
import { useEnvironmentList } from '../queries';
import {
  isDockerEnvironment,
  isKubernetesEnvironment,
  isNomadEnvironment,
} from '../utils';

import { EnvironmentsDatatable } from './EnvironmentsDatatable';
import { useDeleteEnvironmentsMutation } from './useDeleteEnvironmentsMutation';
import { confirmDeleteEnvironments } from './ConfirmDeleteEnvironmentsModal';

export function ListView() {
  const { environments } = useEnvironmentList();
  const constCurrentEnvironmentStore = useStore(environmentStore);
  const deletionMutation = useDeleteEnvironmentsMutation();
  const { trackEvent } = useAnalytics();

  return (
    <>
      <PageHeader
        title="Environments"
        breadcrumbs="Environment management"
        reload
      />

      <EnvironmentsDatatable onRemove={handleRemove} />
    </>
  );

  async function handleRemove(environmentsToDelete: Environment[]) {
    const { confirmed, deleteClusters } =
      (await confirmDeleteEnvironments(environmentsToDelete)) || {};

    if (!confirmed) {
      return;
    }

    // track the number of environments to delete
    trackEvent('delete-environments', {
      category: 'portainer',
      metadata: {
        currentEnvironmentCount: environments.length,
        totalEnvironmentsToDelete: environmentsToDelete.length,
        // Number of clusters to permanently delete by provider
        permanentDeleteCountsByK8sProvider: {
          microk8s: deleteClusters
            ? environmentsToDelete.filter(
                (e) => e.CloudProvider?.Provider === 'microk8s'
              ).length
            : 0,
        },
        // Count of environments to delete by provider
        deleteCountsByK8sProvider: {
          microk8s: environmentsToDelete.filter(
            (e) => e.CloudProvider?.Provider === 'microk8s'
          ).length,
        },
        // Count of environments to delete by platform
        deleteCountsByPlatform: {
          docker: environmentsToDelete.filter((e) =>
            isDockerEnvironment(e.Type)
          ).length,
          kubernetes: environmentsToDelete.filter((e) =>
            isKubernetesEnvironment(e.Type)
          ).length,
          nomad: environmentsToDelete.filter((e) => isNomadEnvironment(e.Type))
            .length,
        },
      },
    });

    deletionMutation.mutate(
      environmentsToDelete.map((e) => {
        // microk8s is the only provider that can delete clusters (for now)
        if (e.CloudProvider?.Provider === 'microk8s') {
          return {
            id: e.Id,
            deleteCluster: deleteClusters,
            name: e.Name,
          };
        }
        return { id: e.Id, deleteCluster: false, name: e.Name };
      }),
      {
        onSuccess() {
          // If the current endpoint was deleted, then clean the endpoint store
          const id = constCurrentEnvironmentStore.environmentId;
          if (environmentsToDelete.some((e) => e.Id === id)) {
            constCurrentEnvironmentStore.clear();
          }
        },
      }
    );
  }
}
