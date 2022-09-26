import { CellProps, Column } from 'react-table';

import { DockerImage } from '@/react/docker/images/types';

export const tags: Column<DockerImage> = {
  Header: 'Tags',
  accessor: 'RepoTags',
  id: 'tags',
  Cell,
  disableFilters: true,
  canHide: true,
  Filter: () => null,
};

function Cell({ value: repoTags }: CellProps<DockerImage, string[]>) {
  return (
    <>
      {repoTags.map((tag, idx) => (
        <span key={idx} className="label label-primary image-tag" title={tag}>
          {tag}
        </span>
      ))}
    </>
  );
}
