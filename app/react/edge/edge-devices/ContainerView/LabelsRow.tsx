import { DockerContainer } from '@/react/docker/containers/types';

import { DetailsRow } from '@@/DetailsTable/DetailsRow';

interface Props {
  labels: DockerContainer['Labels'] | undefined;
}

export function LabelsRow({ labels = {} }: Props) {
  const labelList = Object.entries(labels);

  if (labelList.length === 0) {
    return null;
  }

  return (
    <DetailsRow label="Labels">
      <table className="table table-bordered table-condensed">
        <tbody>
          {labelList.map(([label, value]) => (
            <tr key={label}>
              <td>{label}</td>
              <td>{value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </DetailsRow>
  );
}
