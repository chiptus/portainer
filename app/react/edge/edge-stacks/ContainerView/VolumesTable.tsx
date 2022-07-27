import { DockerContainer } from '@/react/docker/containers/types';

import { DetailsTable } from '@@/DetailsTable';
import { Widget, WidgetTitle, WidgetBody } from '@@/Widget';

interface Props {
  mounts: DockerContainer['Mounts'] | undefined;
}

export function VolumesTable({ mounts = [] }: Props) {
  if (mounts.length === 0) {
    return null;
  }

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle title="Volumes" icon="fa-hdd" />
          <WidgetBody className="!p-0">
            <DetailsTable headers={['Host/Volume', 'Path in container']}>
              {mounts.map((mount) => {
                const name = mount.Type === 'bind' ? mount.Source : mount.Name;

                return (
                  <DetailsTable.Row label={name || mount.Source} key={name}>
                    {mount.Destination}
                  </DetailsTable.Row>
                );
              })}
            </DetailsTable>
          </WidgetBody>
        </Widget>
      </div>
    </div>
  );
}
