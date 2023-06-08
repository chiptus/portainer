import { Field, useFormikContext } from 'formik';
import { useMemo, useState } from 'react';
import { AlertCircle, ArrowLeftRight, CheckCircle, Info } from 'lucide-react';
import { partition } from 'lodash';

import { Credential } from '@/react/portainer/settings/sharedCredentials/types';
import { CustomTemplate } from '@/react/portainer/custom-templates/types';

import { FormControl } from '@@/form-components/FormControl';
import { TextTip } from '@@/Tip/TextTip';
import { LoadingButton } from '@@/buttons';
import { Select, Option } from '@@/form-components/Input/Select';
import { FormError } from '@@/form-components/FormError';

import { CredentialsField } from '../../WizardKaaS/shared/CredentialsField';
import { TestSSHConnectionResponse } from '../../WizardKaaS/types';
import { useSetAvailableOption } from '../../WizardKaaS/useSetAvailableOption';
import { MoreSettingsSection } from '../../shared/MoreSettingsSection';
import { NameField } from '../../shared/NameField';
import { useTestSSHConnection } from '../service';
import { K8sInstallFormValues, Microk8sK8sVersion } from '../types';
import { formatMicrok8sPayload } from '../utils';
import { CustomTemplateSelector } from '../../shared/CustomTemplateSelector';

import { AddOnOption, Microk8sAddOnSelector } from './AddonSelector';
import { NodeAddressErrors, NodeAddressInput } from './NodeAddressInput';
import { Microk8sActions } from './Microk8sActions';

type Props = {
  credentials: Credential[];
  customTemplates: CustomTemplate[];
  isSubmitting: boolean;
  setIsSSHTestSuccessful: (isSuccessful: boolean) => void;
  isSSHTestSuccessful?: boolean;
};

const addonOptions = [
  'metrics-server',
  'ingress',
  'cert-manager',
  'host-access',
  'hostpath-storage',
  'gpu',
  'observability',
  'registry',
];

const microk8sOptions: Option<Microk8sK8sVersion>[] = [
  { label: 'latest/stable', value: 'latest/stable' },
  { label: '1.27/stable', value: '1.27/stable' },
  { label: '1.26/stable', value: '1.26/stable' },
  { label: '1.25/stable', value: '1.25/stable' },
  { label: '1.24/stable', value: '1.24/stable' },
];

export function Microk8sCreateClusterForm({
  credentials,
  customTemplates,
  isSubmitting,
  isSSHTestSuccessful,
  setIsSSHTestSuccessful,
}: Props) {
  const { values, setFieldValue, errors } =
    useFormikContext<K8sInstallFormValues>();
  const testSSHConnectionMutation = useTestSSHConnection();
  const [failedAddressResults, setFailedAddressResults] =
    useState<TestSSHConnectionResponse>([]);
  const [addressResults, setAddressResults] =
    useState<TestSSHConnectionResponse>([]);
  const [testedAddressList, setTestedAddressList] = useState<string[]>([]);
  const [isTestConnectionLoading, setIsTestConnectionLoading] = useState(false);

  const { credentialId } = values;

  const credentialOptions = useMemo(
    () =>
      credentials.map((c) => ({
        value: c.id,
        label: c.name,
      })),
    [credentials]
  );

  // ensure the form values are valid when the options change
  useSetAvailableOption(credentialOptions, credentialId, 'credentialId');

  const isExperimentalVersion = useMemo(
    () => values.microk8s.kubernetesVersion !== '1.24/stable',
    [values.microk8s.kubernetesVersion]
  );

  return (
    <>
      <TextTip
        color="blue"
        icon={Info}
        className="mt-2 !items-start [&>svg]:mt-0.5"
      >
        <p>
          This will allow you to install MicroK8s Kubernetes to your own
          existing nodes, and will then deploy the Portainer agent to it.
        </p>
        <p>
          Only nodes with an operating system of Ubuntu 20.04 LTS and above are
          supported (although other distributions and versions may work).
        </p>
        <p>
          Tested with MicroK8s versions 1.24 to 1.26. Note that if you select
          &apos;latest/stable&apos; and it has moved on from those tested, we
          cannot guarantee support.
        </p>
      </TextTip>
      <NameField
        tooltip="Name of the cluster and environment."
        placeholder="e.g. my-cluster-name"
      />

      <CredentialsField credentials={credentials} />

      <FormControl
        label="Node IP list"
        tooltip="For 3+ node clusters, the first 3 nodes entered are set up as control plane. For 1 or 2 node clusters, the first node entered is set up as control plane. Any remaining are set up as worker nodes."
        inputId="microk8s-nodeIps"
        errors={errors.microk8s?.nodeIPs}
        required
        // reduce the bottom gap so that the test connection button is closer to the input (but still below the front end validation errors)
        className="!mb-0 [&>div>.help-block>p]:!mb-0 [&>div>.help-block]:!mb-0"
      >
        <TextTip
          color="blue"
          className="mt-2 !items-start [&>svg]:mt-0.5"
          icon={Info}
        >
          Add a list of comma or line separated IP addresses. You can also add
          IP ranges by separating with a hyphen e.g. 192.168.1.1 - 192.168.1.10,
          192.168.100.1
        </TextTip>
        <Field
          name="microk8s.nodeIPs"
          as={NodeAddressInput}
          type="text"
          data-cy="microk8sCreateForm-nodeIpsInput"
          id="microk8s-nodeIps"
        />
      </FormControl>
      <FormControl label="" className="[&>label]:!pt-0">
        <div className="mt-2 flex flex-col">
          <LoadingButton
            size="small"
            color="default"
            className="!ml-0 w-min"
            isLoading={isTestConnectionLoading}
            icon={ArrowLeftRight}
            loadingText="Testing connections..."
            onClick={async () => {
              setIsTestConnectionLoading(true); // set this manually, because the mutation is also triggered when submitting is started
              await handleTestConnection();
            }}
            disabled={!!errors.microk8s?.nodeIPs}
            type="button"
          >
            Test connections
          </LoadingButton>
          {isSSHTestSuccessful !== undefined && ( // dont show the text tip if provisioning is started and the test is successful
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
          )}
        </div>
      </FormControl>
      <FormControl
        label="Kubernetes version"
        tooltip="Kubernetes version running on the cluster."
        inputId="microk8s-kubernetesVersion"
        errors={errors.microk8s?.kubernetesVersion}
      >
        <Field
          name="microk8s.kubernetesVersion"
          as={Select}
          id="microk8s-kubernetesVersion"
          data-cy="microk8sCreateForm-kubernetesVersionSelect"
          options={microk8sOptions}
        />
        {isExperimentalVersion && (
          <FormError className="mt-1 !items-start [&>svg]:mt-0.5">
            MicroK8s 1.25 and 1.26 can have an issue running metrics server in
            certain circumstances, which may require a patch to the metric
            server deployment to work around. See{' '}
            <a
              href="https://docs.portainer.io/admin/environments/add/kube-create/microk8s#a-note-about-microk8s-versions"
              target="_blank"
              rel="noreferrer noopener"
            >
              Portainer documentation
            </a>{' '}
            for version-specific issues.
          </FormError>
        )}
      </FormControl>

      <FormControl
        label="Addons"
        tooltip={
          <>
            You may specify{' '}
            <a
              href="https://microk8s.io/docs/addons"
              target="_blank"
              rel="noreferrer noopener"
            >
              addons
            </a>{' '}
            to be automatically installed in your cluster. The following addons
            will also be installed by default: dns, ha-cluster, helm, helm3 and
            rbac.
          </>
        }
        inputId="microk8s-addons"
      >
        <Microk8sAddOnSelector
          value={values.microk8s.addons.map((name) => ({ name }))}
          options={addonOptions.map((name) => ({ name }))}
          onChange={(value: AddOnOption[]) => {
            setFieldValue(
              'microk8s.addons',
              value.map((option) => option.name)
            );
          }}
        />
      </FormControl>

      <MoreSettingsSection>
        <TextTip color="blue" className="mb-4">
          Metadata is only assigned to the environment in Portainer, i.e. the
          group and tags are not assigned to the cluster at the cloud provider
          end.
        </TextTip>
        <CustomTemplateSelector customTemplates={customTemplates} />
      </MoreSettingsSection>

      <Microk8sActions
        isSubmitting={isSubmitting}
        handleTestConnection={handleTestConnection}
        testedAddressList={testedAddressList}
        addressResults={addressResults}
        isSSHTestSuccessful={isSSHTestSuccessful}
      />
    </>
  );

  // handleTestConnection tests the SSH connection to the nodes and returns a boolean indicating whether the test was successful
  function handleTestConnection(): Promise<[boolean, number]> {
    return new Promise((resolve) => {
      const payload = formatMicrok8sPayload(values);
      testSSHConnectionMutation.mutate(
        { payload },
        {
          onSuccess: (addressResults) => {
            const [failedAddressResults, successfulAddressResults] = partition(
              addressResults,
              (result) => result.error
            );
            const isTestSuccessful = failedAddressResults.length === 0;
            // update the component state with the results of the test
            setAddressResults(addressResults);
            setTestedAddressList(values.microk8s.nodeIPs);
            setIsSSHTestSuccessful(isTestSuccessful);
            setFailedAddressResults(failedAddressResults);
            // resolve with the results of the test, and the number of successful addresses
            resolve([isTestSuccessful, successfulAddressResults.length]);
          },
          onError: () => {
            setTestedAddressList(values.microk8s.nodeIPs);
            resolve([false, 0]);
          },
          onSettled: () => {
            setIsTestConnectionLoading(false);
          },
        }
      );
    });
  }
}
