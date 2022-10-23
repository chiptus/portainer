import { find } from 'lodash';
import { Box, Cpu, Database, List } from 'react-feather';
import { useCurrentStateAndParams } from '@uirouter/react';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { humanize } from '@/portainer/filters/filters';
import { useTags } from '@/portainer/tags/queries';
import { ContainerStatus } from '@/react/docker/DashboardView/ContainerStatus';
import { ImagesTotalSize } from '@/react/docker/DashboardView/ImagesTotalSize';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';

import { Widget } from '@@/Widget';
import { DashboardGrid } from '@@/DashboardItem/DashboardGrid';
import { DashboardItem } from '@@/DashboardItem';
import { Icon } from '@@/Icon';
import { Link } from '@@/Link';

import { NoSnapshotAvailablePanel } from '../../components/NoSnapshotAvailablePanel';
import { EdgeDeviceViewsHeader } from '../../components/EdgeDeviceViewsHeader';

export function DashboardView() {
  const {
    params: { environmentId },
  } = useCurrentStateAndParams();

  if (!environmentId) {
    throw new Error('Missing environmentId parameter');
  }

  const environmentQuery = useEnvironment(environmentId);
  const snapshotQuery = useDockerSnapshot(environmentId);
  const tagsQuery = useTags();

  const { data: environment } = environmentQuery;
  const { data: snapshot } = snapshotQuery;

  if (!environment || !tagsQuery.tags) {
    return null;
  }

  const { tags } = tagsQuery;

  const tagsString = environment.TagIds.length
    ? environment.TagIds.map((id) => find(tags, { Id: id }))
        .filter(Boolean)
        .join(', ')
    : '-';

  const totalCpu = environment.Snapshots[0]
    ? environment.Snapshots[0].TotalCPU
    : '-';
  const totalMemory = environment.Snapshots[0]
    ? humanize(environment.Snapshots[0].TotalMemory)
    : '-';

  const breadcrumbs = [
    { label: 'Edge Devices', link: 'edge.devices' },
    { label: environment.Name },
    { label: 'Dashboard' },
  ];

  if (!snapshot) {
    return (
      <>
        <EdgeDeviceViewsHeader
          title="Dashboard"
          breadcrumbs={breadcrumbs}
          environment={environment}
        />

        <NoSnapshotAvailablePanel />
      </>
    );
  }

  const { Containers: containers, Images: images, Volumes: volumes } = snapshot;

  const imagesTotalSize = images.reduce(
    (res, image) => res + image.VirtualSize,
    0
  );

  return (
    <>
      <EdgeDeviceViewsHeader
        title="Dashboard"
        breadcrumbs={breadcrumbs}
        environment={environment}
        snapshot={snapshot}
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <Widget.Title icon="svg-tachometer" title="Environment info" />
            <Widget.Body className="no-padding !px-5">
              <table className="table">
                <tbody>
                  <tr>
                    <td>Environment</td>
                    <td>
                      {environment.Name}{' '}
                      <span className="small text-muted space-left">
                        <Icon icon={Cpu} /> {totalCpu}{' '}
                        <Icon icon="svg-memory" inline /> {totalMemory}
                      </span>
                      <span className="small text-muted">
                        {' '}
                        - Agent {environment.Agent.Version}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td>Tags</td>
                    <td>{tagsString}</td>
                  </tr>
                </tbody>
              </table>
            </Widget.Body>
          </Widget>
        </div>
      </div>

      <div className="mx-4">
        <DashboardGrid>
          <Link
            to="edge.browse.containers"
            params={{ environmentId }}
            className="no-link"
          >
            <DashboardItem
              icon={Box}
              type="Container"
              value={containers.length}
            >
              <ContainerStatus containers={containers} />
            </DashboardItem>
          </Link>
          <Link
            to="edge.browse.images"
            params={{ environmentId }}
            className="no-link"
          >
            <DashboardItem icon={List} type="Image" value={images.length}>
              <ImagesTotalSize imagesTotalSize={imagesTotalSize} />
            </DashboardItem>
          </Link>
          <Link
            to="edge.browse.volumes"
            params={{ environmentId }}
            className="no-link"
          >
            <DashboardItem
              icon={Database}
              type="Volume"
              value={volumes.length}
            />
          </Link>
        </DashboardGrid>
      </div>
    </>
  );
}
