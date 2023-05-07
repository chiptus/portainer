import { useState } from 'react';
import { useFormikContext } from 'formik';
import { Plus } from 'lucide-react';
import { isEqual } from 'lodash';

import { usePublicSettings } from '@/react/portainer/settings/queries';
import { trackEvent } from '@/angulartics.matomo/analytics-services';

import { LoadingButton } from '@@/buttons/LoadingButton';
import { FormSection } from '@@/form-components/FormSection';

import { K8sDistributionType, K8sInstallFormValues } from '../types';
import { TestSSHConnectionResponse } from '../../WizardKaaS/types';

interface Props {
  isSubmitting: boolean;
  handleTestConnection: () => Promise<[boolean, number]>;
  testedAddressList: string[];
  addressResults: TestSSHConnectionResponse;
  isSSHTestSuccessful: boolean | undefined;
}

// Microk8sActions exists to handle custom logic around form submission with node testing
export function Microk8sActions({
  isSubmitting,
  handleTestConnection,
  testedAddressList,
  addressResults,
  isSSHTestSuccessful,
}: Props) {
  const settingsQuery = usePublicSettings();
  const [isTestingConnection, setIsTestingConnection] = useState(false);
  const { values, isValid, submitForm } =
    useFormikContext<K8sInstallFormValues>();

  const isCurrentValuesTested = isEqual(
    testedAddressList.filter((ip) => ip),
    values.microk8s.nodeIPs.filter((ip) => ip).map((ip) => ip.trim())
  );
  const isCurrentValuesFailed =
    isSSHTestSuccessful === false && isCurrentValuesTested;

  const disableProvision = isCurrentValuesFailed || !isValid;

  return (
    <FormSection title="Actions">
      <div className="mb-3 flex w-full flex-wrap gap-2">
        <LoadingButton
          disabled={disableProvision}
          isLoading={isSubmitting || isTestingConnection}
          type="button"
          loadingText="Provision in progress..."
          icon={Plus}
          className="!ml-0"
          onClick={async () => {
            // if already tested and successful, submit form
            if (isCurrentValuesTested && isSSHTestSuccessful) {
              sendAnalytics(addressResults.length);
              submitForm();
              return;
            }
            // otherwise, test connection and submit form if the test is successful
            try {
              setIsTestingConnection(true);
              const [isTestConnectionSuccessful, nodeCount] =
                await handleTestConnection();
              if (isTestConnectionSuccessful) {
                submitForm();
                if (settingsQuery.data?.EnableTelemetry) {
                  sendAnalytics(nodeCount);
                }
              }
            } finally {
              setIsTestingConnection(false);
            }
          }}
        >
          Provision environment
        </LoadingButton>
      </div>
    </FormSection>
  );

  function sendAnalytics(nodeCount: number) {
    trackEvent('portainer-endpoint-creation', {
      category: 'portainer',
      metadata: { type: 'agent', platform: 'kubernetes' },
    });
    trackEvent('create-k8s-cluster', {
      category: 'kubernetes',
      metadata: {
        provider: K8sDistributionType.MICROK8S,
        addons: values.microk8s.addons,
        nodeCount,
        customTemplateUsed: values.meta.customTemplateId !== 0,
      },
    });
  }
}
