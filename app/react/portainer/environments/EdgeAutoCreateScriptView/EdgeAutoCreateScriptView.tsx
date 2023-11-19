import { withLimitToBE } from '@/react/hooks/useLimitToBE';

import { PageHeader } from '@@/PageHeader';

import { AutomaticEdgeEnvCreation } from './AutomaticEdgeEnvCreation';

export const EdgeAutoCreateScriptViewWrapper = withLimitToBE(
  EdgeAutoCreateScriptView
);

function EdgeAutoCreateScriptView() {
  return (
    <>
      <PageHeader
        title="Auto onboarding script creation"
        breadcrumbs={[
          { label: 'Environments', link: 'portainer.endpoints' },
          'Auto onboarding script creation',
        ]}
        reload
      />

      <div className="mx-3">
        <AutomaticEdgeEnvCreation />
      </div>
    </>
  );
}
