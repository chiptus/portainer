import { useCurrentStateAndParams } from '@uirouter/react';

import { react2angular } from '@/react-tools/react2angular';
import { useUser } from '@/portainer/hooks/useUser';

import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { PageHeader } from '@@/PageHeader';
import { Widget, WidgetBody } from '@@/Widget';

import { useGitCredential } from '../gitCredential.service';

import { EditGitCredentialForm } from './EditGitCredentialForm';

export default function EditGitCredentialView() {
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
                <EditGitCredentialForm gitCredential={gitCredential} />
              )}
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}

export const editGitCredentialViewAngular = react2angular(
  EditGitCredentialView,
  []
);
