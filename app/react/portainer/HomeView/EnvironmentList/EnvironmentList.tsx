import { ReactNode, useEffect, useState } from 'react';
import clsx from 'clsx';
import { HardDrive, RefreshCcw } from 'lucide-react';
import _ from 'lodash';
import { useStore } from 'zustand';

import { usePaginationLimitState } from '@/react/hooks/usePaginationLimitState';
import {
  Environment,
  EnvironmentType,
  PlatformType,
  EdgeTypes,
} from '@/react/portainer/environments/types';
import { useDebouncedValue } from '@/react/hooks/useDebouncedValue';
import {
  refetchIfAnyOffline,
  useEnvironmentList,
} from '@/react/portainer/environments/queries/useEnvironmentList';
import { useGroups } from '@/react/portainer/environments/environment-groups/queries';
import { useUser } from '@/react/hooks/useUser';
import { environmentStore } from '@/react/hooks/current-environment-store';
import { useListSelection } from '@/react/hooks/useListSelection';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { EnvironmentsQueryParams } from '@/react/portainer/environments/environment.service';

import { TableFooter } from '@@/datatables/TableFooter';
import { TableActions, TableContainer, TableTitle } from '@@/datatables';
import {
  FilterSearchBar,
  useSearchBarState,
} from '@@/datatables/FilterSearchBar';
import { Button } from '@@/buttons';
import { PaginationControls } from '@@/PaginationControls';

import { ConnectionType } from './types';
import { EnvironmentItem } from './EnvironmentItem';
import { KubeconfigButton } from './KubeconfigButton';
import { NoEnvironmentsInfoPanel } from './NoEnvironmentsInfoPanel';
import { UpdateBadge } from './UpdateBadge';
import styles from './EnvironmentList.module.css';
import { EnvironmentListFilters } from './EnvironmentListFilters';
import { useFiltersStore } from './filters-store';

interface Props {
  onClickBrowse(environment: Environment): void;
  onRefresh(): void;
}

const storageKey = 'home_endpoints';

export function EnvironmentList({ onClickBrowse, onRefresh }: Props) {
  const [selectedItems, handleChangeSelect] = useListSelection<Environment>(
    [],
    (a, b) => a.Id === b.Id
  );

  const { isAdmin } = useUser();
  const { environmentId: currentEnvironmentId } = useStore(environmentStore);
  const [searchBarValue, setSearchBarValue] = useSearchBarState(storageKey);
  const [pageLimit, setPageLimit] = usePaginationLimitState(storageKey);
  const [page, setPage] = useState(1);
  const debouncedTextFilter = useDebouncedValue(searchBarValue);
  const { value: filterValue } = useFiltersStore();

  const groupsQuery = useGroups();

  const environmentsQueryParams: EnvironmentsQueryParams = {
    types: getTypes(filterValue.platformTypes, filterValue.connectionTypes),
    search: debouncedTextFilter,
    status: filterValue.status,
    tagIds: filterValue.tagIds,
    tagsPartialMatch: true,
    groupIds: filterValue.groupIds,
    provisioned: true,
    agentVersions: filterValue.agentVersions,
    updateInformation: isBE,
  };

  const {
    isLoading,
    environments,
    totalCount,
    totalAvailable,
    updateAvailable,
  } = useEnvironmentList(
    {
      page,
      pageLimit,
      sort: filterValue.sort,
      order: filterValue.sortDesc ? 'desc' : 'asc',
      ...environmentsQueryParams,
    },
    refetchIfAnyOffline
  );

  useEffect(() => {
    setPage(1);
  }, [searchBarValue]);

  return (
    <>
      {totalAvailable === 0 && <NoEnvironmentsInfoPanel isAdmin={isAdmin} />}
      <div className="row">
        <div className="col-sm-12">
          <TableContainer>
            <TableTitle icon={HardDrive} label="Environments">
              {updateAvailable && <UpdateBadge />}
            </TableTitle>

            <TableActions className={styles.actionBar}>
              <div className={styles.description}>
                Click on an environment to manage
              </div>
              <div className={styles.actionButton}>
                <div className={styles.refreshButton}>
                  {isAdmin && (
                    <Button
                      onClick={onRefresh}
                      data-cy="home-refreshEndpointsButton"
                      size="medium"
                      color="secondary"
                      className={clsx(
                        'vertical-center !ml-0',
                        styles.refreshEnvironmentsButton
                      )}
                    >
                      <RefreshCcw
                        className="lucide icon-sm icon-white"
                        aria-hidden="true"
                      />
                      Refresh
                    </Button>
                  )}
                </div>
                <div className={styles.kubeconfigButton}>
                  <KubeconfigButton
                    environments={environments}
                    envQueryParams={{
                      ...environmentsQueryParams,
                      sort: filterValue.sort,
                      order: filterValue.sortDesc ? 'desc' : 'asc',
                    }}
                    selectedItems={selectedItems}
                  />
                </div>
                <div className={clsx(styles.filterSearchbar, 'ml-3')}>
                  <FilterSearchBar
                    value={searchBarValue}
                    onChange={setSearchBarValue}
                    placeholder="Search by name, group, tag, status, URL..."
                    data-cy="home-endpointsSearchInput"
                  />
                </div>
              </div>
            </TableActions>
            <EnvironmentListFilters />
            <div className="blocklist" data-cy="home-endpointList">
              {renderItems(
                isLoading,
                totalCount,
                environments.map((env) => (
                  <EnvironmentItem
                    key={env.Id}
                    environment={env}
                    groupName={
                      groupsQuery.data?.find((g) => g.Id === env.GroupId)?.Name
                    }
                    onClickBrowse={() => onClickBrowse(env)}
                    isActive={env.Id === currentEnvironmentId}
                    isSelected={selectedItems.some(
                      (selectedEnv) => selectedEnv.Id === env.Id
                    )}
                    onSelect={(selected) => handleChangeSelect(env, selected)}
                  />
                ))
              )}
            </div>

            <TableFooter>
              <PaginationControls
                showAll={totalCount <= 100}
                pageLimit={pageLimit}
                page={page}
                onPageChange={setPage}
                totalCount={totalCount}
                onPageLimitChange={setPageLimit}
              />
            </TableFooter>
          </TableContainer>
        </div>
      </div>
    </>
  );

  function renderItems(
    isLoading: boolean,
    totalCount: number,

    items: ReactNode
  ) {
    if (isLoading) {
      return (
        <div className="text-center text-muted" data-cy="home-loadingEndpoints">
          Loading...
        </div>
      );
    }

    if (!totalCount) {
      return (
        <div className="text-center text-muted" data-cy="home-noEndpoints">
          No environments available.
        </div>
      );
    }

    return items;
  }
}

function getTypes(
  platformTypes: PlatformType[],
  connectionTypes: ConnectionType[]
) {
  if (platformTypes.length === 0 && connectionTypes.length === 0) {
    return [];
  }

  const typesByPlatform = {
    [PlatformType.Docker]: [
      EnvironmentType.Docker,
      EnvironmentType.AgentOnDocker,
      EnvironmentType.EdgeAgentOnDocker,
    ],
    [PlatformType.Azure]: [EnvironmentType.Azure],
    [PlatformType.Kubernetes]: [
      EnvironmentType.KubernetesLocal,
      EnvironmentType.AgentOnKubernetes,
      EnvironmentType.EdgeAgentOnKubernetes,
    ],
    [PlatformType.Nomad]: [EnvironmentType.EdgeAgentOnNomad],
  };

  const typesByConnection = {
    [ConnectionType.API]: [
      EnvironmentType.Azure,
      EnvironmentType.KubernetesLocal,
      EnvironmentType.Docker,
    ],
    [ConnectionType.Agent]: [
      EnvironmentType.AgentOnDocker,
      EnvironmentType.AgentOnKubernetes,
    ],
    [ConnectionType.EdgeAgent]: EdgeTypes,
    [ConnectionType.EdgeDevice]: EdgeTypes,
  };

  const selectedTypesByPlatform = platformTypes.flatMap(
    (platformType) => typesByPlatform[platformType]
  );
  const selectedTypesByConnection = connectionTypes.flatMap(
    (connectionType) => typesByConnection[connectionType]
  );

  if (selectedTypesByPlatform.length === 0) {
    return selectedTypesByConnection;
  }

  if (selectedTypesByConnection.length === 0) {
    return selectedTypesByPlatform;
  }

  return _.intersection(selectedTypesByConnection, selectedTypesByPlatform);
}
