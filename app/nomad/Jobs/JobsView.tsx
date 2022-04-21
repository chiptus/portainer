import { react2angular } from '@/react-tools/react2angular';
import { PageHeader } from '@/portainer/components/PageHeader';
import { TableSettingsProvider } from '@/portainer/components/datatables/components/useTableSettings';
import { useEnvironmentId } from '@/portainer/hooks/useEnvironmentId';
import { useSnapshot } from '@/nomad/hooks/useSnapshot';

import { JobsDatatable } from './JobsDatatable';

export function JobsView() {
  const environmentId = useEnvironmentId();
  const { query, invalidateQuery } = useSnapshot(environmentId);

  const data = query.data?.Jobs || [];

  const defaultSettings = {
    autoRefreshRate: 10,
    pageSize: 10,
    sortBy: { id: 'name', desc: false },
  };

  async function onReload() {
    invalidateQuery();
  }

  return (
    <>
      <PageHeader
        title="Nomad Job list"
        breadcrumbs={[{ label: 'Nomad Jobs' }]}
        reload
        loading={query.isFetching}
        onReload={onReload}
      />

      <div className="row">
        <div className="col-sm-12">
          <TableSettingsProvider defaults={defaultSettings} storageKey="jobs">
            <JobsDatatable
              jobs={data}
              refreshData={onReload}
              isLoading={query.isLoading}
            />
          </TableSettingsProvider>
        </div>
      </div>
    </>
  );
}

export const JobsViewAngular = react2angular(JobsView, []);