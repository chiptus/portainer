import { useState } from 'react';

import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { PageHeader } from '@@/PageHeader';
import { Widget, WidgetBody } from '@@/Widget';

import { CredentialType } from '../types';
import { CloudProviderSelector } from '../components/SharedCredentialSelector';

import { CredentialsForm } from './CredentialsForm';

export function CreateCredentialView() {
  const [selectedProvider, setSelectedProvider] = useState<CredentialType>(
    CredentialType.CIVO
  );
  return (
    <>
      <PageHeader
        title="Add shared credentials"
        breadcrumbs={[
          { label: 'Settings', link: 'portainer.settings' },
          {
            label: 'Shared credentials',
            link: 'portainer.settings.sharedcredentials',
          },
          { label: 'Add credentials' },
        ]}
        reload
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetBody>
              <FormSectionTitle>Provider</FormSectionTitle>
              <div className="form-horizontal">
                <CloudProviderSelector
                  value={selectedProvider}
                  onChange={(provider: CredentialType) => {
                    setSelectedProvider(provider);
                  }}
                />
              </div>
              <CredentialsForm
                credentialType={selectedProvider}
                routeOnSuccess="portainer.settings.sharedcredentials"
              />
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}
