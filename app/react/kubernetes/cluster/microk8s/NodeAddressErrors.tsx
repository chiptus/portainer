import { useState } from 'react';

import { TestSSHConnectionResponse } from '@/react/kubernetes/cluster/microk8s/microk8s.service';

import { Button } from '@@/buttons';

type NodeAddressErrorsProps = {
  failedAddressResults: TestSSHConnectionResponse;
  addressResults: TestSSHConnectionResponse;
};

export function NodeAddressErrors({
  addressResults,
  failedAddressResults,
}: NodeAddressErrorsProps) {
  const [showMoreFailedResults, setShowMoreFailedResults] = useState(false);

  return (
    <>
      <b>{`${failedAddressResults.length} out of ${addressResults.length} nodes have errors:`}</b>
      <ul className="mb-0">
        {failedAddressResults.map(({ address, error }, index) => (
          <li
            key={index}
            className={!showMoreFailedResults && index >= 2 ? 'hidden' : ''}
          >{`${address}: ${error}`}</li>
        ))}
      </ul>
      {failedAddressResults.length > 2 && (
        <Button
          color="link"
          className="!ml-0 !pl-0"
          onClick={() => setShowMoreFailedResults(!showMoreFailedResults)}
        >
          {showMoreFailedResults ? 'See less...' : 'See more...'}
        </Button>
      )}
    </>
  );
}
