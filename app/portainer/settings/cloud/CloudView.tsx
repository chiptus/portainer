import { PageHeader } from '@/portainer/components/PageHeader';
import { react2angular } from '@/react-tools/react2angular';

import { CredentialsDatatableContainer } from './CloudCredentialsDatatable/CredentialsDatatableContainer';

export function CloudView() {
  return (
    <>
      <PageHeader
        title="Cloud Settings"
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

export const CloudViewAngular = react2angular(CloudView, []);
