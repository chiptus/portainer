import { useCurrentStateAndParams } from '@uirouter/react';

import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { PageHeader } from '@@/PageHeader';
import { Widget, WidgetBody } from '@@/Widget';

import { useCloudCredential } from '../cloudSettings.service';

import { EditCredentialForm } from './EditCredentialForm';

export function EditCredentialView() {
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
