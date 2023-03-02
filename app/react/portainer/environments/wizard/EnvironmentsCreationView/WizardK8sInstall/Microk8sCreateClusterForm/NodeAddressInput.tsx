import { useState } from 'react';
import { useFormikContext } from 'formik';

import { TextArea } from '@@/form-components/Input/Textarea';
import { Button } from '@@/buttons';

import { K8sInstallFormValues } from '../types';
import { TestSSHConnectionResponse } from '../../WizardKaaS/types';

// this input is part of a formik form. When the text area is changed, the textbox string is separated into an array of strings by new line and comma separators.
// This array is set as the 'microk8s.nodeIPs' formik value
export function NodeAddressInput() {
  const { values, setFieldValue } = useFormikContext<K8sInstallFormValues>();
  return (
    <div>
      <TextArea
        className="min-h-[150px] resize-y"
        value={values.microk8s.nodeIPs.join('\n')} // display the text area as a string with each new ip address/entry on a new line
        onChange={(e) => {
          const nodeIpArrayFromString = e.target.value.split('\n');
          setFieldValue('microk8s.nodeIPs', nodeIpArrayFromString);
        }}
      />
    </div>
  );
}

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
