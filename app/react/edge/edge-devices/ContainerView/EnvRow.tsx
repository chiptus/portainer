import { getPairKey, getPairValue } from '@/portainer/filters/filters';
import { DockerContainerSnapshot } from '@/react/docker/snapshots/types';

import { DetailsRow } from '@@/DetailsTable/DetailsRow';

interface Props {
  env: DockerContainerSnapshot['Env'] | undefined;
}

export function EnvRow({ env = [] }: Props) {
  if (env.length === 0) {
    return null;
  }

  return (
    <DetailsRow label="ENV">
      <table className="table-bordered table-condensed table">
        <tbody>
          {env.map((entry) => (
            <tr key={getPairKey(entry, '=')}>
              <td>{getPairKey(entry, '=')}</td>
              <td>{getPairValue(entry, '=')}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </DetailsRow>
  );
}
