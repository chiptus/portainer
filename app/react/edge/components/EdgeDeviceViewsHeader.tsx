import { Environment } from '@/react/portainer/environments/types';

import { PageHeader, PageHeaderProps } from '@@/PageHeader';

import { DockerSnapshotPanel } from './SnapshotPanel';

type Props = {
  environment: Environment;
} & Omit<PageHeaderProps, 'reload'>;

export function EdgeDeviceViewsHeader({
  environment,

  title,
  breadcrumbs,
  ...props
}: Props) {
  return (
    // eslint-disable-next-line react/jsx-props-no-spreading
    <PageHeader {...props} title={title} breadcrumbs={breadcrumbs} reload>
      <DockerSnapshotPanel environment={environment} />
    </PageHeader>
  );
}
