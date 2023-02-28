import { useCurrentStateAndParams } from '@uirouter/react';

import { useUser } from '@/react/hooks/useUser';

import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { PageHeader } from '@@/PageHeader';
import { Widget, WidgetBody } from '@@/Widget';

import { useGitCredential } from '../git-credentials.service';

import { EditGitCredentialsForm } from './EditGitCredentialsForm';

export function EditGitCredentialsView() {
  const currentUser = useUser();
  const { params } = useCurrentStateAndParams();
  const gitCredentialQuery = useGitCredential(currentUser.user.Id, params.id);
  const gitCredential = gitCredentialQuery.data;
  return (
    <>
      <PageHeader
        title="Edit Git Credential"
        breadcrumbs={[
          { label: 'My account', link: 'portainer.account' },
          { label: 'Edit credential' },
        ]}
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetBody loading={false}>
              <FormSectionTitle>Edit Git Credential</FormSectionTitle>
              {gitCredential && (
                <EditGitCredentialsForm gitCredential={gitCredential} />
              )}
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}
