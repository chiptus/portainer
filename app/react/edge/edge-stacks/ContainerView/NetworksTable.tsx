import { SummaryNetworkSettings } from '@/react/docker/containers/types/response';

import { Widget, WidgetTitle, WidgetBody } from '@@/Widget';
import { DetailsTable } from '@@/DetailsTable';

interface Props {
  networks: SummaryNetworkSettings['Networks'] | undefined;
}

export function NetworksTable({ networks = {} }: Props) {
  const networksList = Object.entries(networks);

  if (networksList.length === 0) {
    return null;
  }

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle title="Connected networks" icon="fa-sitemap" />
          <WidgetBody className="!p-0">
            <DetailsTable
              headers={['Network', 'IP Address', 'Gateway', 'MAC Address']}
            >
              {networksList.map(([name, network]) => (
                <tr key={name}>
                  <td>{name}</td>
                  <td>{network?.IPAddress}</td>
                  <td>{network?.Gateway || '-'}</td>
                  <td>{network?.MacAddress}</td>
                </tr>
              ))}
            </DetailsTable>
          </WidgetBody>
        </Widget>
      </div>
    </div>
  );
}
