import _ from 'lodash';
import { Tag, Activity } from 'lucide-react';

import {
  isoDateFromTimestamp,
  stripProtocol,
} from '@/portainer/filters/filters';
import {
  type Environment,
  PlatformType,
} from '@/react/portainer/environments/types';
import {
  getDashboardRoute,
  getPlatformType,
  isEdgeEnvironment,
} from '@/react/portainer/environments/utils';
import type { TagId } from '@/portainer/tags/types';
import { useTags } from '@/portainer/tags/queries';

import { EdgeIndicator } from '@@/EdgeIndicator';
import { EnvironmentStatusBadge } from '@@/EnvironmentStatusBadge';
import { Link } from '@@/Link';

import { EnvironmentIcon } from './EnvironmentIcon';
import { EnvironmentStats } from './EnvironmentStats';
import { EngineVersion } from './EngineVersion';
import { EnvironmentTypeTag } from './EnvironmentTypeTag';
import { EnvironmentBrowseButtons } from './EnvironmentBrowseButtons';
import { EditButtons } from './EditButtons';
import { AgentDetails } from './AgentDetails';

interface Props {
  environment: Environment;
  groupName?: string;
  onClickBrowse(): void;
  isActive: boolean;
}

export function EnvironmentItem({
  environment,
  onClickBrowse,
  groupName,
  isActive,
}: Props) {
  const isEdge = isEdgeEnvironment(environment.Type);

  const snapshotTime = getSnapshotTime(environment);

  const tags = useEnvironmentTagNames(environment.TagIds);

  return (
    <div className="relative">
      <Link
        to={getDashboardRoute(environment)}
        params={{
          endpointId: environment.Id,
          environmentId: environment.Id,
        }}
        className="no-link"
      >
        <button
          className="blocklist-item flex items-stretch overflow-hidden min-h-[110px] bg-transparent w-full !m-0 !pr-56"
          onClick={onClickBrowse}
          type="button"
        >
          <div className="ml-2 self-center flex justify-center">
            <EnvironmentIcon type={environment.Type} />
          </div>
          <div className="ml-3 mr-auto flex justify-center gap-3 flex-col items-start">
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
              <span className="font-bold">{environment.Name}</span>
              {isEdge ? (
                <EdgeIndicator environment={environment} showLastCheckInDate />
              ) : (
                <>
                  <EnvironmentStatusBadge status={environment.Status} />
                  {snapshotTime && (
                    <span
                      className="small text-muted vertical-center gap-1"
                      title="Last snapshot time"
                    >
                      <Activity className="icon icon-sm" aria-hidden="true" />
                      {snapshotTime}
                    </span>
                  )}
                </>
              )}
              <EngineVersion environment={environment} />
              {!isEdge && (
                <span className="text-muted small vertical-center">
                  {stripProtocol(environment.URL)}
                </span>
              )}
            </div>
            <div className="small text-muted flex flex-wrap items-center gap-x-4 gap-y-2">
              {groupName && (
                <span className="font-semibold">
                  <span>Group: </span>
                  <span>{groupName}</span>
                </span>
              )}
              <span className="vertical-center gap-1">
                <Tag className="icon icon-sm" aria-hidden="true" />
                {tags}
              </span>
              <EnvironmentTypeTag environment={environment} />
              <AgentDetails environment={environment} />
            </div>
            <EnvironmentStats environment={environment} />
          </div>
        </button>
      </Link>
      {/* 
      Buttons are extracted out of the main button because it causes errors with react and accessibility issues
      see https://stackoverflow.com/questions/66409964/warning-validatedomnesting-a-cannot-appear-as-a-descendant-of-a
      */}
      <div className="absolute inset-y-0 right-0 flex justify-end w-56">
        <div className="py-3 flex items-center">
          <EnvironmentBrowseButtons
            environment={environment}
            onClickBrowse={onClickBrowse}
            isActive={isActive}
          />
        </div>
        <EditButtons environment={environment} />
      </div>
    </div>
  );
}

function useEnvironmentTagNames(tagIds?: TagId[]) {
  const { tags, isLoading } = useTags((tags) => {
    if (!tagIds) {
      return [];
    }
    return _.compact(
      tagIds
        .map((id) => tags.find((tag) => tag.ID === id))
        .map((tag) => tag?.Name)
    );
  });

  if (tags && tags.length > 0) {
    return tags.join(', ');
  }

  if (isLoading) {
    return 'Loading tags...';
  }

  return 'No tags';
}

function getSnapshotTime(environment: Environment) {
  const platform = getPlatformType(environment.Type);

  switch (platform) {
    case PlatformType.Docker:
      return environment.Snapshots.length > 0
        ? isoDateFromTimestamp(environment.Snapshots[0].Time)
        : null;
    case PlatformType.Kubernetes:
      return environment.Kubernetes.Snapshots &&
        environment.Kubernetes.Snapshots.length > 0
        ? isoDateFromTimestamp(environment.Kubernetes.Snapshots[0].Time)
        : null;
    default:
      return null;
  }
}
