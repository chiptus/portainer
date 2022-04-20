import { TableSettingsMenuAutoRefresh } from '@/portainer/components/datatables/components/TableSettingsMenuAutoRefresh';
import { useTableSettings } from '@/portainer/components/datatables/components/useTableSettings';

import { JobsTableSettings } from './types';

export function JobsDatatableSettings() {
  const { settings, setTableSettings } = useTableSettings<JobsTableSettings>();

  return (
    <TableSettingsMenuAutoRefresh
      value={settings.autoRefreshRate}
      onChange={handleRefreshRateChange}
    />
  );

  function handleRefreshRateChange(autoRefreshRate: number) {
    setTableSettings({ autoRefreshRate });
  }
}
