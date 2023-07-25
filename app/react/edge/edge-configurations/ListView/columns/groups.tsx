import _ from 'lodash';
import { CellContext } from '@tanstack/react-table';

import { EdgeGroup } from '@/react/edge/edge-groups/types';
import { useEdgeGroups } from '@/react/edge/edge-groups/queries/useEdgeGroups';

import { Link } from '@@/Link';

import { EdgeConfiguration } from '../../types';

import { columnHelper } from './helper';

export const groups = columnHelper.accessor('edgeGroupIDs', {
  header: 'Edge Groups',
  cell: GroupsCell,
});

function GroupsCell({
  getValue,
}: CellContext<EdgeConfiguration, Array<EdgeGroup['Id']>>) {
  const groupsIds = getValue();
  const groupsQuery = useEdgeGroups();

  const groups = _.compact(
    groupsIds.map((id) => groupsQuery.data?.find((g) => g.Id === id))
  );

  const lastItem = groups.length - 1;
  return (
    <>
      {groups.map((g, idx) => (
        <span key={idx}>
          <Link to="edge.groups.edit" params={{ groupId: g.Id }} title={g.Name}>
            {g.Name}
          </Link>
          {idx !== lastItem ? ', ' : ''}
        </span>
      ))}
    </>
  );
}
