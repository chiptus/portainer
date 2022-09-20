import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { react2angular } from '@/react-tools/react2angular';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';

import { PageHeader } from '@@/PageHeader';

import { CredentialsDatatableContainer } from './CloudCredentialsDatatable/CredentialsDatatableContainer';

export function CloudView() {
  return (
    <>
      <PageHeader
        title="Cloud settings"
        breadcrumbs={[
          { label: 'Settings', link: 'portainer.settings' },
          { label: 'Cloud' },
        ]}
      />

      <div className="row">
        <div className="col-sm-12">
          <CredentialsDatatableContainer />
        </div>
      </div>
    </>
  );
}

export const CloudViewAngular = react2angular(
  withUIRouter(withReactQuery(withCurrentUser(CloudView))),
  []
);
