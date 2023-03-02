import { PageHeader } from '@@/PageHeader';

import { CredentialsDatatable } from './CloudCredentialsDatatable/CredentialsDatatable';

export function CloudView() {
  return (
    <>
      <PageHeader
        title="Shared credentials"
        breadcrumbs={[
          { label: 'Settings', link: 'portainer.settings' },
          { label: 'Shared credentials' },
        ]}
      />

      <CredentialsDatatable />
    </>
  );
}
