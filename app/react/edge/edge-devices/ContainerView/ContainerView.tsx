import { useCurrentStateAndParams } from '@uirouter/react';

import { useEnvironment } from '@/portainer/environments/queries';
import { useDockerSnapshotContainer } from '@/react/docker/queries/useDockerSnapshotContainer';
import { isoDateFromTimestamp } from '@/portainer/filters/filters';

import { PageHeader } from '@@/PageHeader';
import { DetailsTable } from '@@/DetailsTable';
import { Widget, WidgetTitle, WidgetBody } from '@@/Widget';

import { NetworksTable } from './NetworksTable';
import { VolumesTable } from './VolumesTable';
import { LabelsRow } from './LabelsRow';
import { StatusBadge } from './StatusBadge';

export function ContainerView() {
  const {
    params: { environmentId, containerId },
  } = useCurrentStateAndParams();

  const environmentQuery = useEnvironment(environmentId);

  const containerQuery = useDockerSnapshotContainer(environmentId, containerId);

  if (!environmentId || !containerId) {
    throw new Error('Missing environmentId, stackId or containerId parameters');
  }

  if (!environmentQuery.data || !containerQuery.data) {
    return null;
  }

  const { data: environment } = environmentQuery;
  const { data: container } = containerQuery;

  const name = container.Names[0].substring(1);

  return (
    <>
      <PageHeader
        title="Containers"
        breadcrumbs={[
          { label: 'Edge Devices', link: 'edge.devices' },
          { label: environment.Name },
          { label: 'Containers', link: '^' },
          { label: name },
        ]}
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetTitle title="Container status" icon="fa-server" />
            <WidgetBody className="!p-0">
              <DetailsTable>
                <DetailsTable.Row label="ID">{container.Id}</DetailsTable.Row>
                <DetailsTable.Row label="Name">{name}</DetailsTable.Row>
                <DetailsTable.Row label="Status">
                  <StatusBadge
                    status={container.Status}
                    state={container.State}
                  />
                </DetailsTable.Row>
                <DetailsTable.Row label="Created">
                  {isoDateFromTimestamp(container.Created)}
                </DetailsTable.Row>
              </DetailsTable>
            </WidgetBody>
          </Widget>
        </div>
      </div>

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetTitle title="Container details" icon="fa-server" />
            <WidgetBody className="!p-0">
              <DetailsTable>
                <DetailsTable.Row label="Image">
                  {container.Image}
                </DetailsTable.Row>
                <DetailsTable.Row label="CMD">
                  {container.Command}
                </DetailsTable.Row>
                <LabelsRow labels={container.Labels} />
              </DetailsTable>
            </WidgetBody>
          </Widget>
        </div>
      </div>

      <VolumesTable mounts={container.Mounts} />

      <NetworksTable networks={container.NetworkSettings?.Networks} />
    </>
  );
}
