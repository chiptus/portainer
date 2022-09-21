import { useEnvironment } from '@/portainer/environments/queries';
import { useEnvironmentId } from '@/portainer/hooks/useEnvironmentId';

import { Datatable } from '@@/datatables';
import { Button } from '@@/buttons';

import { useIngressControllerClassMap } from '../queries';

import { RowProvider } from './RowContext';
import { useColumns } from './columns';
import { createStore } from './datatable-store';

const useStore = createStore('ingressClasses');

export function IngressClassDatatable() {
  const envId = useEnvironmentId();
  const environmentQuery = useEnvironment(envId);
  const controllerClassMapQuery = useIngressControllerClassMap();
  const settings = useStore();
  const columns = useColumns();

  if (environmentQuery.isLoading) {
    return <div>Loading environment...</div>;
  }

  if (environmentQuery.isError) {
    return <div>Error getting environments...</div>;
  }

  function renderTableActions() {
    return (
      <>
        <Button color="dangerlight" onClick={() => console.log('none')}>
          Disallow all
        </Button>
        <Button color="primary" onClick={() => console.log('all')}>
          Allow all
        </Button>
      </>
    );
  }

  return (
    <div className="-mx-[15px]">
      {environmentQuery.data && (
        <RowProvider context={{ environment: environmentQuery.data }}>
          <Datatable
            dataset={controllerClassMapQuery.data || []}
            storageKey="ingressClasses"
            columns={columns}
            settingsStore={settings}
            isLoading={controllerClassMapQuery.isLoading}
            emptyContentLabel="No supported ingress controllers found"
            titleOptions={{
              icon: 'database',
              title: 'Ingress controllers',
              featherIcon: true,
            }}
            getRowId={(row) => row.Name + row.Type}
            renderTableActions={renderTableActions}
          />
        </RowProvider>
      )}
    </div>
  );
}
