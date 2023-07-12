import { useMemo, useState } from 'react';
import { useFormikContext } from 'formik';
import { Plus } from 'lucide-react';
import { isEqual } from 'lodash';

import { TestSSHConnectionResponse } from '@/react/kubernetes/cluster/microk8s/microk8s.service';
import { useAnalytics } from '@/react/hooks/useAnalytics';
import { K8sDistributionType } from '@/react/portainer/environments/types';

import { LoadingButton } from '@@/buttons/LoadingButton';
import { FormSection } from '@@/form-components/FormSection';

import { K8sInstallFormValues } from '../types';

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
  const { trackEvent } = useAnalytics();
  const [isTestingConnection, setIsTestingConnection] = useState(false);
  const { values, isValid, submitForm } =
    useFormikContext<K8sInstallFormValues>();

  const isCurrentValuesTested = useMemo(() => {
    const allNodeIPs = [
      ...values.microk8s.masterNodes,
      ...values.microk8s.workerNodes,
    ];
    return isEqual(
      testedAddressList.filter((ip) => ip),
      allNodeIPs.filter((ip) => ip).map((ip) => ip.trim())
    );
  }, [
    testedAddressList,
    values.microk8s.masterNodes,
    values.microk8s.workerNodes,
  ]);

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
          onClick={async () => handleProvision()}
        >
          Provision environment
        </LoadingButton>
      </div>
    </FormSection>
  );

  async function handleProvision() {
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
        sendAnalytics(nodeCount);
      }
    } finally {
      setIsTestingConnection(false);
    }
  }

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
