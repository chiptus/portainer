import { useState } from 'react';

import { react2angular } from '@/react-tools/react2angular';

import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { PageHeader } from '@@/PageHeader';
import { Widget, WidgetBody } from '@@/Widget';

import { KaasProvider } from '../types';
import { CloudProviderSelector } from '../components/CloudProviderSelector';

import { CredentialsForm } from './CredentialsForm';

export default function CreateCredentialView() {
  const [selectedProvider, setSelectedProvider] = useState<KaasProvider>(
    KaasProvider.CIVO
  );
  return (
    <>
      <PageHeader
        title="Add cloud provider credentials"
        breadcrumbs={[
          { label: 'Settings', link: 'portainer.settings' },
          { label: 'Cloud', link: 'portainer.settings.cloud' },
          { label: 'Add credentials' },
        ]}
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetBody>
              <FormSectionTitle>Cloud service provider</FormSectionTitle>
              <CloudProviderSelector
                value={selectedProvider}
                onChange={(provider: KaasProvider) => {
                  setSelectedProvider(provider);
                }}
              />
              <CredentialsForm
                selectedProvider={selectedProvider}
                routeOnSuccess="portainer.settings.cloud"
              />
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}

export const CreateCredentialViewAngular = react2angular(
  CreateCredentialView,
  []
);
