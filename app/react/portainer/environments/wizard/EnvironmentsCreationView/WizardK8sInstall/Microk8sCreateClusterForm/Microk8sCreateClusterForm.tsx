import { Field, useFormikContext } from 'formik';
import { useMemo, useState } from 'react';
import { ArrowLeftRight, Info } from 'lucide-react';
import { partition } from 'lodash';

import { Credential } from '@/react/portainer/settings/sharedCredentials/types';
import { CustomTemplate } from '@/react/portainer/custom-templates/types';
import {
  TestSSHConnectionResponse,
  useTestSSHConnection,
} from '@/react/kubernetes/cluster/microk8s/microk8s.service';
import { NodeAddressInput } from '@/react/kubernetes/cluster/microk8s/NodeAddressInput';
import { formatNodeIPs } from '@/react/kubernetes/cluster/microk8s/utils';
import { NodeAddressTestResults } from '@/react/kubernetes/cluster/microk8s/NodeAddressTestResults';

import { FormControl } from '@@/form-components/FormControl';
import { TextTip } from '@@/Tip/TextTip';
import { LoadingButton } from '@@/buttons';
import { Select } from '@@/form-components/Input/Select';

import { CredentialsField } from '../../WizardKaaS/shared/CredentialsField';
import { useSetAvailableOption } from '../../WizardKaaS/useSetAvailableOption';
import { MoreSettingsSection } from '../../shared/MoreSettingsSection';
import { NameField } from '../../shared/NameField';
import { K8sInstallFormValues } from '../types';
import { useMicroK8sOptions } from '../queries';
import { CustomTemplateSelector } from '../../shared/CustomTemplateSelector';

import { AddOnOption, Microk8sAddOnSelector } from './AddonSelector';
import { Microk8sActions } from './Microk8sActions';

type Props = {
  credentials: Credential[];
  customTemplates: CustomTemplate[];
  isSubmitting: boolean;
  setIsSSHTestSuccessful: (isSuccessful: boolean) => void;
  isSSHTestSuccessful?: boolean;
};

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

  const { credentialId, microk8s } = values;

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

  const microk8sOptionsQuery = useMicroK8sOptions();
  const microk8sOptions = microk8sOptionsQuery.data;
  const kubernetesVersions = useMemo(
    () => microk8sOptions?.kubernetesVersions || [],
    [microk8sOptions?.kubernetesVersions]
  );
  useSetAvailableOption(
    kubernetesVersions,
    microk8s.kubernetesVersion,
    'kubernetesVersion'
  );

  const addonOptions: AddOnOption[] = useMemo(() => {
    const addonOptions: AddOnOption[] = [];
    microk8sOptions?.availableAddons.forEach((a) => {
      const kubeVersion = parseFloat(microk8s.kubernetesVersion.split('/')[0]);
      const versionAvailableFrom = parseFloat(a.versionAvailableFrom);
      if (kubeVersion >= versionAvailableFrom) {
        addonOptions.push({ name: a.label, type: a.type });
      }
    });

    addonOptions.sort(
      (a, b) => b.type.localeCompare(a.type) || a.name.localeCompare(b.name)
    );
    return addonOptions;
  }, [microk8sOptions?.availableAddons, microk8s.kubernetesVersion]);

  return (
    <>
      <TextTip color="blue" icon={Info} className="mt-2">
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
        label="Control plane nodes"
        tooltip="Control plane nodes manage cluster state and workload scheduling on worker nodes. For high availability, use 3 nodes (or 5 for greater reliability)."
        inputId="microk8s-masterNodesToAdd"
        errors={errors.microk8s?.masterNodes}
        required
      >
        <TextTip
          color="blue"
          className="mt-2 !items-start [&>svg]:mt-0.5"
          icon={Info}
        >
          <p>
            Add a list of comma or line separated IP addresses. You can also add
            IP ranges by separating with a hyphen e.g. 192.168.1.1 -
            192.168.1.10, 192.168.100.1
          </p>
          <p>
            Your nodes must be internet routable from this Portainer instance,
            and you must ensure ports 22, 16443 and 30778 are open to them. WSL
            will not typically meet this.
          </p>
        </TextTip>
        <Field
          name="microk8s.masterNodes"
          as={NodeAddressInput}
          type="text"
          data-cy="microk8sCreateForm-controlPlaneNodesInput"
          id="microk8s-nodeIps"
          nodeIPValues={values.microk8s.masterNodes}
          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => {
            const nodeIpArrayByLine = e.target.value.split('\n');
            setFieldValue('microk8s.masterNodes', nodeIpArrayByLine);
          }}
        />
      </FormControl>
      <FormControl
        label="Worker nodes"
        tooltip="Worker nodes execute tasks assigned by the control plane nodes and handle the execution of containers and workloads to keep your applications running smoothly."
        inputId="workerNodesToAdd"
        errors={errors.microk8s?.workerNodes}
        // reduce the bottom gap so that the test connection button is closer to the input (but still below the front end validation errors)
        className="!mb-0 [&>div>.help-block>p]:!mb-0 [&>div>.help-block]:!mb-0"
      >
        <Field
          name="microk8s.workerNodes"
          as={NodeAddressInput}
          type="text"
          data-cy="microk8sCreateForm-workerNodesInput"
          id="controlPlaneNodes"
          nodeIPValues={values.microk8s.workerNodes}
          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => {
            const nodeIpArrayByLine = e.target.value.split('\n');
            setFieldValue('microk8s.workerNodes', nodeIpArrayByLine);
          }}
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
            disabled={
              !!errors.microk8s?.masterNodes || !!errors.microk8s?.workerNodes
            }
            type="button"
          >
            Test connections
          </LoadingButton>
          <NodeAddressTestResults
            failedAddressResults={failedAddressResults}
            addressResults={addressResults}
            isSSHTestSuccessful={isSSHTestSuccessful}
          />
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
          options={kubernetesVersions}
        />
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
            will also be installed by default: community, dns, ha-cluster, helm,
            helm3 and rbac.
          </>
        }
        inputId="microk8s-addons"
      >
        <Microk8sAddOnSelector
          value={values.microk8s.addons}
          options={addonOptions}
          onChange={(value: AddOnOption[]) => {
            setFieldValue('microk8s.addons', value);
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
      const combinedNodeIPs = formatNodeIPs([
        ...values.microk8s.masterNodes,
        ...values.microk8s.workerNodes,
      ]);
      testSSHConnectionMutation.mutate(
        {
          nodeIPs: combinedNodeIPs,
          credentialID: values.credentialId,
        },
        {
          onSuccess: (addressResults) => {
            const [failedAddressResults, successfulAddressResults] = partition(
              addressResults,
              (result) => result.error
            );
            const isTestSuccessful = failedAddressResults.length === 0;
            // update the component state with the results of the test
            setAddressResults(addressResults);
            setTestedAddressList(combinedNodeIPs);
            setIsSSHTestSuccessful(isTestSuccessful);
            setFailedAddressResults(failedAddressResults);
            // resolve with the results of the test, and the number of successful addresses
            resolve([isTestSuccessful, successfulAddressResults.length]);
          },
          onError: () => {
            setTestedAddressList(combinedNodeIPs);
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
