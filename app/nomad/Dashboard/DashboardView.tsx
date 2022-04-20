import { react2angular } from '@/react-tools/react2angular';
import { DashboardItem } from '@/portainer/components/Dashboard/DashboardItem';
import { Widget, WidgetTitle, WidgetBody } from '@/portainer/components/widget';
import { PageHeader } from '@/portainer/components/PageHeader';
import { useEnvironmentId } from '@/portainer/hooks/useEnvironmentId';
import { useSnapshot } from '@/nomad/hooks/useSnapshot';

import { RunningStatus } from './RunningStatus';

export function DashboardView() {
  const environmentId = useEnvironmentId();
  const { query } = useSnapshot(environmentId);

  const snapshot = query.data;

  return (
    <>
      <PageHeader
        title="Dashboard"
        breadcrumbs={[{ label: 'Environment summary' }]}
      />
      {query.isLoading && (
        <div className="text-center" style={{ marginTop: '30%' }}>
          Connecting to the Edge environment...
          <i className="fa fa-cog fa-spin space-left" />
        </div>
      )}

      {!query.isLoading && (
        <div className="row">
          <div className="col-sm-12">
            {/* title */}
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
                      <td>{snapshot?.NodeCount ?? '-'}</td>
                    </tr>
                  </tbody>
                </table>
              </WidgetBody>
            </Widget>
          </div>
          <div className="row">
            {/* jobs */}
            <DashboardItem
              value={snapshot?.JobCount}
              icon="fa fa-th-list"
              type="Nomad Jobs"
            />
            {/* groups */}
            <DashboardItem
              value={snapshot?.GroupCount}
              icon="fa fa-list-alt"
              type="Groups"
            />
            {/* tasks */}
            <DashboardItem
              value={snapshot?.TaskCount}
              icon="fa fa-cubes"
              type="Tasks"
            >
              {/* running status of tasks */}
              <RunningStatus
                running={snapshot.RunningTaskCount}
                stopped={snapshot.TaskCount - snapshot.RunningTaskCount}
              />
            </DashboardItem>
          </div>
        </div>
      )}
    </>
  );
}

export const NomadDashboardAngular = react2angular(DashboardView, []);
