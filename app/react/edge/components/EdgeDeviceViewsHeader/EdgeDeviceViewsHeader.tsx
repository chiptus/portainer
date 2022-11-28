import { PropsWithChildren } from 'react';
import { ChevronDown, History } from 'lucide-react';
import { addSeconds } from 'date-fns';
import { Menu, MenuButton, MenuPopover } from '@reach/menu-button';
import { get } from 'lodash';

import { Environment } from '@/react/portainer/environments/types';
import { DockerSnapshot } from '@/react/docker/snapshots/types';
import { useSettings } from '@/react/portainer/settings/queries';

import { EnvironmentStatusBadge } from '@@/EnvironmentStatusBadge';
import { PageHeaderProps, PageHeader } from '@@/PageHeader';
import { DetailsTable } from '@@/DetailsTable';
import { Icon } from '@@/Icon';

import { DateColumn } from './DateColumn';
import { IntervalColumn } from './IntervalColumn';

type Props = {
  environment: Environment;
  snapshot?: DockerSnapshot;
} & Omit<PageHeaderProps, 'reload'>;

export function EdgeDeviceViewsHeader({
  environment,
  snapshot,
  title,
  breadcrumbs,
  ...props
}: Props) {
  return (
    // eslint-disable-next-line react/jsx-props-no-spreading
    <PageHeader {...props} title={title} breadcrumbs={breadcrumbs} reload>
      {snapshot && (
        <SnapshotPanel environment={environment} snapshot={snapshot} />
      )}
    </PageHeader>
  );
}

interface SnapshotPanelProps {
  environment: Environment;
  snapshot: DockerSnapshot;
}

function SnapshotPanel({ environment, snapshot }: SnapshotPanelProps) {
  const { data: defaultSnapshotInterval } = useSettings((settings): number =>
    get(settings, 'Edge.SnapshotInterval', 0)
  );

  const snapshotInterval =
    environment.Edge.SnapshotInterval !== -1
      ? environment.Edge.SnapshotInterval
      : defaultSnapshotInterval || 0;

  const lastSnapshotDate = new Date(snapshot.SnapshotTime);
  const now = new Date();
  const nextSnapshotDate = addSeconds(lastSnapshotDate, snapshotInterval);
  const snapshotIntervalEnd = addSeconds(now, snapshotInterval);

  const snapshotDisabled = snapshotInterval === 0;

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
                <EnvironmentStatusBadge status={environment.Status} />
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
