import { PropsWithChildren } from 'react';
import { ChevronDown, History } from 'lucide-react';
import { addSeconds } from 'date-fns';
import { Menu, MenuButton, MenuPopover } from '@reach/menu-button';
import { get } from 'lodash';

import {
  Environment,
  EnvironmentStatus,
} from '@/react/portainer/environments/types';
import { useSettings } from '@/react/portainer/settings/queries';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import {
  isDockerEnvironment,
  isEdgeAsync,
} from '@/react/portainer/environments/utils';

import { EnvironmentStatusBadge } from '@@/EnvironmentStatusBadge';
import { DetailsTable } from '@@/DetailsTable';
import { Icon } from '@@/Icon';

import { DateColumn } from './DateColumn';
import { IntervalColumn } from './IntervalColumn';

export function DockerSnapshotPanel({
  environment,
}: {
  environment: Environment;
}) {
  const isDocker = isDockerEnvironment(environment.Type);
  const isBrowsingSnapshot = isEdgeAsync(environment);

  const snapshotQuery = useDockerSnapshot(environment.Id, {
    enabled: isDocker && isBrowsingSnapshot,
  });
  const snapshot = snapshotQuery.data;
  if (!isDocker || !isBrowsingSnapshot || !snapshot) {
    return null;
  }

  return (
    <SnapshotPanel
      environment={environment}
      lastSnapshotTime={snapshot.Info.SystemTime}
    />
  );
}

interface SnapshotPanelProps {
  environment: Environment;
  lastSnapshotTime: string;
}

function SnapshotPanel({ environment, lastSnapshotTime }: SnapshotPanelProps) {
  const { data: defaultSnapshotInterval } = useSettings((settings): number =>
    get(settings, 'Edge.SnapshotInterval', 0)
  );

  const snapshotInterval =
    environment.Edge.SnapshotInterval !== -1
      ? environment.Edge.SnapshotInterval
      : defaultSnapshotInterval || 0;

  const lastSnapshotDate = new Date(lastSnapshotTime);
  const now = new Date();
  const nextSnapshotDate = addSeconds(lastSnapshotDate, snapshotInterval);
  const snapshotIntervalEnd = addSeconds(now, snapshotInterval);

  const snapshotDisabled = snapshotInterval === 0;

  const status = environment.Heartbeat
    ? EnvironmentStatus.Up
    : EnvironmentStatus.Down;
  return (
    <div className="ml-5">
      <Menu>
        <MenuButton className="form-control flex items-center">
          <Icon icon={History} className="!mr-2" />
          <span>Browsing snapshot</span>
          <ChevronDown className="ml-5" />
        </MenuButton>
        <MenuPopover className="dropdown-menu">
          <div className="tableMenu py-0">
            <DetailsTable className="!m-0">
              <Row label="Last updated">
                <DateColumn date={lastSnapshotDate} />
                <IntervalColumn
                  start={lastSnapshotDate}
                  end={now}
                  pattern="($ ago)"
                  muted
                />
              </Row>
              <Row label="Next update">
                {environment.Heartbeat && (
                  <>
                    <DateColumn
                      date={nextSnapshotDate}
                      snapshotDisabled={snapshotDisabled}
                    />
                    <IntervalColumn
                      start={now}
                      end={nextSnapshotDate}
                      pattern="(in $)"
                      snapshotDisabled={snapshotDisabled}
                      muted
                    />
                  </>
                )}
                {!environment.Heartbeat && (
                  <span>No update planned as environment is offline</span>
                )}
              </Row>
              {!snapshotDisabled && (
                <Row label="Snapshot interval">
                  <IntervalColumn
                    start={now}
                    end={snapshotIntervalEnd}
                    pattern="Every $"
                  />
                </Row>
              )}
              <Row label="Environment status">
                <EnvironmentStatusBadge status={status} />
              </Row>
            </DetailsTable>
          </div>
        </MenuPopover>
      </Menu>
    </div>
  );
}

interface RowProps {
  label: string;
}

function Row({ label, children }: PropsWithChildren<RowProps>) {
  return (
    <DetailsTable.Row
      label={label}
      className="!h-1"
      colClassName="!py-0 !border-none"
    >
      {children}
    </DetailsTable.Row>
  );
}
