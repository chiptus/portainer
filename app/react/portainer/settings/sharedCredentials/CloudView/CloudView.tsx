import { useRouter } from '@uirouter/react';
import { useEffect } from 'react';

import { useCurrentUser } from '@/react/hooks/useUser';

import { PageHeader } from '@@/PageHeader';

import { CredentialsDatatable } from './CloudCredentialsDatatable/CredentialsDatatable';

export function CloudView() {
  const { isAdmin } = useCurrentUser();
  const router = useRouter();

  useEffect(() => {
    if (!isAdmin) {
      router.stateService.go('portainer.home');
    }
  }, [isAdmin, router.stateService]);

  return (
    <>
      <PageHeader
        title="Shared credentials"
        breadcrumbs={[
          { label: 'Settings', link: 'portainer.settings' },
          { label: 'Shared credentials' },
        ]}
        reload
      />

      <CredentialsDatatable />
    </>
  );
}
