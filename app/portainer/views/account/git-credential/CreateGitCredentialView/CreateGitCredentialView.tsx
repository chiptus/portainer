import { react2angular } from '@/react-tools/react2angular';

import { PageHeader } from '@@/PageHeader';
import { Widget, WidgetBody } from '@@/Widget';

import { CreateGitCredentialForm } from './CreateGitCredentialForm';

export default function CreateGitCredentialView() {
  return (
    <>
      <PageHeader
        title="Create git authentication"
        breadcrumbs={[
          { label: 'My account', link: 'portainer.account' },
          { label: 'Create git authentication' },
        ]}
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetBody>
              <CreateGitCredentialForm routeOnSuccess="portainer.account" />
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}

export const createGitCredentialViewAngular = react2angular(
  CreateGitCredentialView,
  []
);
