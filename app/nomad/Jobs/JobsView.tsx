import { react2angular } from '@/react-tools/react2angular';
import { useCurrentEnvironmentSnapshot } from '@/portainer/hooks/useCurrentEnvironmentSnapshot';

import { PageHeader } from '@@/PageHeader';
import { TableSettingsProvider } from '@@/datatables/useTableSettings';

import { JobsDatatable } from './JobsDatatable';

export function JobsView() {
  const { isLoading, snapshot, triggerSnapshot } =
    useCurrentEnvironmentSnapshot();

  const data = snapshot?.Jobs || [];

  const defaultSettings = {
    autoRefreshRate: 10,
    pageSize: 10,
    sortBy: { id: 'name', desc: false },
  };

  return (
    <>
      <PageHeader
        title="Nomad Job list"
        breadcrumbs={[{ label: 'Nomad Jobs' }]}
        reload
        loading={isLoading}
        onReload={triggerSnapshot}
      />

      <div className="row">
        <div className="col-sm-12">
          <TableSettingsProvider defaults={defaultSettings} storageKey="jobs">
            <JobsDatatable
              jobs={data}
              refreshData={triggerSnapshot}
              isLoading={isLoading}
            />
          </TableSettingsProvider>
        </div>
      </div>
    </>
  );
}

export const JobsViewAngular = react2angular(JobsView, []);
