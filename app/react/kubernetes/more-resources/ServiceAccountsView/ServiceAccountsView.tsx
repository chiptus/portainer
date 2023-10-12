import { PageHeader } from '@@/PageHeader';

import { ServiceAccountsDatatable } from './ServiceAccountsDatatable';

export function ServiceAccountsView() {
  return (
    <>
      <PageHeader
        title="Service Account List"
        breadcrumbs="Service Accounts"
        reload
      />
      <ServiceAccountsDatatable />
    </>
  );
}
