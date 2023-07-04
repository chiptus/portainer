import { PageHeader } from '@@/PageHeader';
import { Widget } from '@@/Widget';

import { AddNodesForm } from './AddNodesForm';

export function NodeCreateView() {
  return (
    <>
      <PageHeader
        title="Create nodes"
        breadcrumbs={[
          { label: 'Cluster information', link: 'kubernetes.cluster' },
          { label: 'Create nodes' },
        ]}
      />
      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <AddNodesForm />
          </Widget>
        </div>
      </div>
    </>
  );
}
