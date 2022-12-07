import _ from 'lodash';

import {
  EnvironmentStatus,
  PlatformType,
} from '@/react/portainer/environments/types';
import { useGroups } from '@/react/portainer/environments/environment-groups/queries';
import { useAgentVersionsList } from '@/react/portainer/environments/queries/useAgentVersionsList';
import { useTags } from '@/portainer/tags/queries';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';

import { HomepageFilter } from './HomepageFilter';
import { SortSelector } from './SortbySelector';
import styles from './EnvironmentListFilters.module.css';
import { ConnectionType, Filters } from './types';

const statusOptions = [
  { value: EnvironmentStatus.Up, label: 'Up' },
  { value: EnvironmentStatus.Down, label: 'Down' },
];

const sortByOptions = ['Name', 'Group', 'Status'].map((v) => ({
  label: v,
  value: v,
}));

export function EnvironmentListFilters({
  value,
  onChange,
}: {
  value: Filters;
  onChange: (value: Filters) => void;
}) {
  const agentVersionsQuery = useAgentVersionsList();
  const agentVersionOptions =
    agentVersionsQuery.data?.map((v) => ({
      label: v,
      value: v,
    })) || [];

  const groupsQuery = useGroups();
  const groupOptions = [...(groupsQuery.data || [])].map(
    ({ Id: value, Name: label }) => ({
      value,
      label,
    })
  );

  const tagsQuery = useTags();
  const tagOptions = [...(tagsQuery.tags || [])].map(
    ({ ID: value, Name: label }) => ({
      value,
      label,
    })
  );

  const connectionTypeOptions = getConnectionTypeOptions(value.platformTypes);
  const platformTypeOptions = getPlatformTypeOptions(value.connectionTypes);

  return (
    <div className={styles.filterContainer}>
      <div className={styles.filterLeft}>
        <HomepageFilter
          filterOptions={platformTypeOptions}
          onChange={(platformTypes) => handleChange({ platformTypes })}
          placeholder="Platform"
          value={value.platformTypes}
        />
      </div>
      <div className={styles.filterLeft}>
        <HomepageFilter
          filterOptions={connectionTypeOptions}
          onChange={(connectionTypes) => handleChange({ connectionTypes })}
          placeholder="Connection Type"
          value={value.connectionTypes}
        />
      </div>
      <div className={styles.filterLeft}>
        <HomepageFilter
          filterOptions={statusOptions}
          onChange={(status) => handleChange({ status })}
          placeholder="Status"
          value={value.status}
        />
      </div>
      <div className={styles.filterLeft}>
        <HomepageFilter
          filterOptions={tagOptions}
          onChange={(tagIds) =>
            handleChange({ tagIds: tagIds.length > 0 ? tagIds : undefined })
          }
          placeholder="Tags"
          value={value.tagIds || []}
        />
      </div>
      <div className={styles.filterLeft}>
        <HomepageFilter
          filterOptions={groupOptions}
          onChange={(groupIds) => handleChange({ groupIds })}
          placeholder="Groups"
          value={value.groupIds}
        />
      </div>
      <div className={styles.filterLeft}>
        <HomepageFilter<string>
          filterOptions={agentVersionOptions}
          onChange={(agentVersions) => handleChange({ agentVersions })}
          placeholder="Agent Version"
          value={value.agentVersions}
        />
      </div>
      <button
        type="button"
        className={styles.clearButton}
        onClick={clearFilter}
      >
        Clear all
      </button>
      <div className={styles.filterRight}>
        <SortSelector
          filterOptions={sortByOptions}
          onChange={sortOnchange}
          onDescendingChange={sortOnDescending}
          placeHolder="Sort By"
          sortByDescending={value.sortDesc}
          value={value.sort}
        />
      </div>
    </div>
  );

  function handleChange(newValue: Partial<Filters>) {
    onChange({ ...value, ...newValue });
  }

  function clearFilter() {
    handleChange({
      platformTypes: [],
      connectionTypes: [],
      status: [],
      tagIds: undefined,
      groupIds: undefined,
      agentVersions: undefined,
      sort: undefined,
      sortDesc: false,
    });
  }

  function sortOnchange(sort?: string) {
    handleChange({ sort });
  }

  function sortOnDescending(sortDesc: boolean) {
    handleChange({ sortDesc });
  }
}

function getConnectionTypeOptions(platformTypes: ReadonlyArray<PlatformType>) {
  const platformTypeConnectionType = {
    [PlatformType.Docker]: [
      ConnectionType.API,
      ConnectionType.Agent,
      ConnectionType.EdgeAgent,
      ConnectionType.EdgeDevice,
    ],
    [PlatformType.Azure]: [ConnectionType.API],
    [PlatformType.Kubernetes]: [
      ConnectionType.Agent,
      ConnectionType.EdgeAgent,
      ConnectionType.EdgeDevice,
    ],
    [PlatformType.Nomad]: [ConnectionType.EdgeAgent, ConnectionType.EdgeDevice],
  };

  const connectionTypesDefaultOptions = [
    { value: ConnectionType.API, label: 'API' },
    { value: ConnectionType.Agent, label: 'Agent' },
    { value: ConnectionType.EdgeAgent, label: 'Edge Agent' },
  ];

  if (platformTypes.length === 0) {
    return connectionTypesDefaultOptions;
  }

  return _.compact(
    _.intersection(
      ...platformTypes.map((p) => platformTypeConnectionType[p])
    ).map((c) => connectionTypesDefaultOptions.find((o) => o.value === c))
  );
}

function getPlatformTypeOptions(
  connectionTypes: ReadonlyArray<ConnectionType>
) {
  const platformDefaultOptions = [
    { value: PlatformType.Docker, label: 'Docker' },
    { value: PlatformType.Azure, label: 'Azure' },
    { value: PlatformType.Kubernetes, label: 'Kubernetes' },
  ];

  if (isBE) {
    platformDefaultOptions.push({
      value: PlatformType.Nomad,
      label: 'Nomad',
    });
  }

  if (connectionTypes.length === 0) {
    return platformDefaultOptions;
  }

  const connectionTypePlatformType = {
    [ConnectionType.API]: [PlatformType.Docker, PlatformType.Azure],
    [ConnectionType.Agent]: [PlatformType.Docker, PlatformType.Kubernetes],
    [ConnectionType.EdgeAgent]: [
      PlatformType.Kubernetes,
      PlatformType.Nomad,
      PlatformType.Docker,
    ],
    [ConnectionType.EdgeDevice]: [
      PlatformType.Nomad,
      PlatformType.Docker,
      PlatformType.Kubernetes,
    ],
  };

  return _.compact(
    _.intersection(
      ...connectionTypes.map((p) => connectionTypePlatformType[p])
    ).map((c) => platformDefaultOptions.find((o) => o.value === c))
  );
}
