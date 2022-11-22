import { PageHeader } from '@@/PageHeader';

import { CredentialsDatatable } from './CloudCredentialsDatatable/CredentialsDatatable';

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

      <CredentialsDatatable />
    </>
  );
}
