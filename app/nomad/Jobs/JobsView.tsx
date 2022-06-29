import { react2angular } from '@/react-tools/react2angular';
import { useJobs } from '@/nomad/hooks/useJobs';
import { useEnvironmentId } from '@/portainer/hooks/useEnvironmentId';

import { PageHeader } from '@@/PageHeader';
import { TableSettingsProvider } from '@@/datatables/useTableSettings';

import { JobsDatatable } from './JobsDatatable';

export function JobsView() {
  const environmentId = useEnvironmentId();
  const jobsQuery = useJobs(environmentId);

  const defaultSettings = {
    autoRefreshRate: 10,
    pageSize: 10,
    sortBy: { id: 'name', desc: false },
  };

  async function reloadData() {
    await jobsQuery.refetch();
  }

  return (
    <>
      <PageHeader
        title="Nomad Job list"
        breadcrumbs={[{ label: 'Nomad Jobs' }]}
        reload
        loading={jobsQuery.isLoading}
        onReload={reloadData}
      />

      <div className="row">
        <div className="col-sm-12">
          <TableSettingsProvider defaults={defaultSettings} storageKey="jobs">
            <JobsDatatable
              jobs={jobsQuery.data || []}
              refreshData={reloadData}
              isLoading={jobsQuery.isLoading}
            />
          </TableSettingsProvider>
        </div>
      </div>
    </>
  );
}

export const JobsViewAngular = react2angular(JobsView, []);
