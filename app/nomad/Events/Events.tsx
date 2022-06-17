import { useCurrentStateAndParams } from '@uirouter/react';

import { react2angular } from '@/react-tools/react2angular';
import { useEnvironmentId } from '@/portainer/hooks/useEnvironmentId';
import { useEvents } from '@/nomad/hooks/useEvents';
import { EventsDatatable } from '@/nomad/Events/datatable/EventsDatatable';
import { NomadEventsList } from '@/nomad/types';

import { TableSettingsProvider } from '@@/datatables/useTableSettings';
import { PageHeader } from '@@/PageHeader';

export function Events() {
  const environmentId = useEnvironmentId();
  const { query, invalidateQuery } = useEvents();
  const {
    params: { jobID, taskName },
  } = useCurrentStateAndParams();

  const breadcrumbs = [
    {
      label: 'Nomad Jobs',
      link: 'nomad.jobs',
      linkParams: { id: environmentId },
    },
    { label: jobID },
    { label: taskName },
    { label: 'Events' },
  ];

  const defaultSettings = {
    pageSize: 10,
    sortBy: {},
  };

  return (
    <>
      {/* header */}
      <PageHeader
        title="Event list"
        breadcrumbs={breadcrumbs}
        reload
        loading={query.isLoading || query.isFetching}
        onReload={invalidateQuery}
      />

      <div className="row">
        <div className="col-sm-12">
          <TableSettingsProvider
            defaults={defaultSettings}
            storageKey="nomad-events"
          >
            {/* events table */}
            <EventsDatatable
              data={(query.data || []) as NomadEventsList}
              isLoading={query.isLoading}
            />
          </TableSettingsProvider>
        </div>
      </div>
    </>
  );
}

export const NomadEventsAngular = react2angular(Events, []);
