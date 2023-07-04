import { AlertCircle, CheckCircle } from 'lucide-react';

import { TestSSHConnectionResponse } from '@/react/kubernetes/cluster/microk8s/microk8s.service';

import { TextTip } from '@@/Tip/TextTip';

import { NodeAddressErrors } from './NodeAddressErrors';

type Props = {
  failedAddressResults: TestSSHConnectionResponse;
  addressResults: TestSSHConnectionResponse;
  isSSHTestSuccessful?: boolean;
};

export function NodeAddressTestResults({
  addressResults,
  failedAddressResults,
  isSSHTestSuccessful,
}: Props) {
  if (isSSHTestSuccessful === undefined) {
    return null;
  }

  return (
    <TextTip
      className="mt-2 !items-start [&>svg]:mt-0.5"
      icon={isSSHTestSuccessful ? CheckCircle : AlertCircle}
      color={isSSHTestSuccessful ? 'green' : 'red'}
    >
      {isSSHTestSuccessful === false ? (
        <NodeAddressErrors
          failedAddressResults={failedAddressResults}
          addressResults={addressResults}
        />
      ) : (
        `${addressResults.length} out of ${addressResults.length} nodes are reachable.`
      )}
    </TextTip>
  );
}
