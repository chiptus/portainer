import { useEnvironmentId } from 'Portainer/hooks/useEnvironmentId';

import { react2angular } from '@/react-tools/react2angular';
import { useDashboard } from '@/nomad/hooks/useDashboard';

import { DashboardItem } from '@@/DashboardItem';
import { Widget, WidgetTitle, WidgetBody } from '@@/Widget';
import { PageHeader } from '@@/PageHeader';

import { RunningStatus } from './RunningStatus';

export function DashboardView() {
  const environmentId = useEnvironmentId();
  const dashboardQuery = useDashboard(environmentId);

  const running = dashboardQuery.data?.RunningTaskCount || 0;
  const stopped = (dashboardQuery.data?.TaskCount || 0) - running;

  return (
    <>
      <PageHeader
        title="Dashboard"
        breadcrumbs={[{ label: 'Environment summary' }]}
      />

      {dashboardQuery.isLoading ? (
        <div className="text-center" style={{ marginTop: '30%' }}>
          Connecting to the Edge environment...
          <i className="fa fa-cog fa-spin space-left" />
        </div>
      ) : (
        <div className="row">
          <div className="col-sm-12">
            {/* cluster info */}
            <Widget>
              <WidgetTitle
                icon="fa-tachometer-alt"
                title="Cluster information"
              />
              <WidgetBody className="no-padding">
                <table className="table">
                  <tbody>
                    <tr>
                      <td>Nodes in the cluster</td>
                      <td>{dashboardQuery.data?.NodeCount ?? '-'}</td>
                    </tr>
                  </tbody>
                </table>
              </WidgetBody>
            </Widget>
          </div>
          <div className="row">
            {/* jobs */}
            <DashboardItem
              value={dashboardQuery.data?.JobCount}
              icon="fa fa-th-list"
              type="Nomad Jobs"
            />

            {/* groups */}
            <DashboardItem
              value={dashboardQuery.data?.GroupCount}
              icon="fa fa-list-alt"
              type="Groups"
            />

            {/* tasks */}
            <DashboardItem
              value={dashboardQuery.data?.TaskCount}
              icon="fa fa-cubes"
              type="Tasks"
            >
              {/* running status of tasks */}
              <RunningStatus running={running} stopped={stopped} />
            </DashboardItem>
          </div>
        </div>
      )}
    </>
  );
}

export const NomadDashboardAngular = react2angular(DashboardView, []);
