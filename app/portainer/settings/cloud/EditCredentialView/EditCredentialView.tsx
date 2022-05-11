import { useCurrentStateAndParams } from '@uirouter/react';

import { FormSectionTitle } from '@/portainer/components/form-components/FormSectionTitle';
import { PageHeader } from '@/portainer/components/PageHeader';
import { Widget, WidgetBody } from '@/portainer/components/widget';
import { react2angular } from '@/react-tools/react2angular';

import { useCloudCredential } from '../cloudSettings.service';

import { EditCredentialForm } from './EditCredentialForm';

export default function EditCredentialView() {
  const { params } = useCurrentStateAndParams();

  const cloudCredentialQuery = useCloudCredential(params.id);
  const credential = cloudCredentialQuery.data;

  return (
    <>
      <PageHeader
        title="Edit Cloud Provider Credentials"
        breadcrumbs={[
          { label: 'Settings', link: 'portainer.settings' },
          { label: 'Cloud', link: 'portainer.settings.cloud' },
          { label: 'Edit Credentials' },
        ]}
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetBody loading={cloudCredentialQuery.isLoading}>
              <FormSectionTitle>Edit Cloud Credentials</FormSectionTitle>
              {credential && <EditCredentialForm credential={credential} />}
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}

export const EditCredentialViewAngular = react2angular(EditCredentialView, []);
