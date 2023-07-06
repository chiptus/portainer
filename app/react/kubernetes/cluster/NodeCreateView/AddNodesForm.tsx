import { ArrowLeftRight, Info } from 'lucide-react';
import { Field, Form, Formik } from 'formik';
import { SchemaOf, TestContext, object } from 'yup';
import { useRouter } from '@uirouter/react';
import { useState } from 'react';
import { isEqual, partition } from 'lodash';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { validateNodeIPList } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/Microk8sCreateClusterForm/validation';
import { notifySuccess } from '@/portainer/services/notifications';
import { K8sDistributionType } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/types';
import { useAnalytics } from '@/react/hooks/useAnalytics';

import { TextTip } from '@@/Tip/TextTip';
import { FormControl } from '@@/form-components/FormControl';
import { WidgetBody } from '@@/Widget';
import { Button, LoadingButton } from '@@/buttons';
import { FormSection } from '@@/form-components/FormSection';

import { useAddNodesMutation, useNodesQuery } from '../HomeView/nodes.service';
import { NodeAddressInput } from '../microk8s/NodeAddressInput';
import { formatNodeIPs } from '../microk8s/utils';
import {
  TestSSHConnectionResponse,
  useTestSSHConnection,
} from '../microk8s/microk8s.service';
import {
  getInternalNodeIpAddress,
  getRole,
} from '../HomeView/NodesDatatable/utils';
import { NodeAddressTestResults } from '../microk8s/NodeAddressTestResults';

import { AddNodesFormValues } from './types';

export function validation(
  existingNodeIPAddresses?: string[]
): SchemaOf<AddNodesFormValues> {
  return object({
    masterNodesToAdd: validateNodeIPList(existingNodeIPAddresses).test(
      'at least one node',
      'At least one control plane or worker node is required',
      atLeastOneNode
    ),
    workerNodesToAdd: validateNodeIPList(existingNodeIPAddresses).test(
      'at least one node',
      'At least one control plane or worker node is required',
      atLeastOneNode
    ),
  });
}

function atLeastOneNode(this: TestContext) {
  const formValues = this.parent as AddNodesFormValues;
  return (
    !!formValues.masterNodesToAdd?.some((node) => node) ||
    !!formValues.workerNodesToAdd?.some((node) => node)
  );
}

const initialValues: AddNodesFormValues = {
  masterNodesToAdd: [''],
  workerNodesToAdd: [''],
};

export function AddNodesForm() {
  const router = useRouter();
  const { trackEvent } = useAnalytics();

  // initialise state
  const [isTestConnectionLoading, setIsTestConnectionLoading] = useState(false);
  const [isTestingConnectionOnSubmit, setIsTestingConnectionOnSubmit] =
    useState(false);
  const [failedAddressResults, setFailedAddressResults] =
    useState<TestSSHConnectionResponse>([]);
  const [addressResults, setAddressResults] =
    useState<TestSSHConnectionResponse>([]);
  const [testedAddressList, setTestedAddressList] = useState<string[]>([]);
  const [isSSHTestSuccessful, setIsSSHTestSuccessful] = useState<boolean>();

  // get queries
  const environmentId = useEnvironmentId();
  const { data: credentialID, ...environmentQuery } = useEnvironment(
    environmentId,
    (environment) => environment?.CloudProvider.CredentialID
  );
  const { data: nodes, ...nodesQuery } = useNodesQuery(environmentId);
  const existingNodeIPAddresses = nodes
    ?.map((node) => getInternalNodeIpAddress(node))
    .filter((ip): ip is string => ip !== undefined);

  // register mutations
  const addNodesMutation = useAddNodesMutation(environmentId);
  const testSSHConnectionMutation = useTestSSHConnection();

  if (nodesQuery.isLoading || environmentQuery.isLoading) {
    return null;
  }

  return (
    <WidgetBody>
      <Formik
        initialValues={initialValues}
        onSubmit={(values: AddNodesFormValues) => {
          const formattedValues = {
            masterNodesToAdd: formatNodeIPs(values.masterNodesToAdd),
            workerNodesToAdd: formatNodeIPs(values.workerNodesToAdd),
          };
          addNodesMutation.mutate(formattedValues, {
            onSuccess: () => {
              notifySuccess(
                'Success',
                'Request to add nodes successfully submitted. This might take a few minutes to complete.'
              );
              router.stateService.go('portainer.kubernetes.cluster');
            },
          });
        }}
        validateOnMount
        validationSchema={() => validation(existingNodeIPAddresses)}
      >
        {({
          errors,
          isValid,
          handleSubmit,
          setFieldValue,
          values,
          submitForm,
          validateField,
          isSubmitting,
        }) => {
          const allNodeIPs = [
            ...values.masterNodesToAdd,
            ...values.workerNodesToAdd,
          ];
          const isCurrentValuesTested = isEqual(
            testedAddressList.filter((ip) => ip),
            allNodeIPs.filter((ip) => ip).map((ip) => ip.trim())
          );
          const isCurrentValuesFailed =
            isSSHTestSuccessful === false && isCurrentValuesTested;
          const disableAddNodes = isCurrentValuesFailed || !isValid;

          return (
            <Form onSubmit={handleSubmit}>
              <FormControl
                label="Control plane nodes"
                tooltip="Control plane nodes manage cluster state and workload scheduling on worker nodes. For high availability, use 3 nodes (or 5 for greater reliability)."
                inputId="masterNodesToAdd"
                errors={errors.masterNodesToAdd}
                className="[&>div>.help-block>p]:!mb-0 [&>div>.help-block]:!mb-0 [&>label]:!pl-0"
              >
                <TextTip color="blue" icon={Info}>
                  <p>
                    Edit your list of comma or line separated IP addresses. You
                    can also include IP ranges by separating with a hyphen e.g.
                    192.168.1.1 - 192.168.1.3, 192.168.1.100.
                  </p>
                  <p>
                    Your nodes must be internet routable from this Portainer
                    instance, and you must ensure ports 22, 16443 and 30778 are
                    open to them. WSL will not typically meet this.
                  </p>
                </TextTip>
                <Field
                  name="masterNodesToAdd"
                  as={NodeAddressInput}
                  type="text"
                  data-cy="microk8sEditForm-controlPlaneNodesInput"
                  id="masterNodesToAdd"
                  nodeIPValues={values.masterNodesToAdd}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => {
                    const nodeIpArrayByLine = e.target.value.split('\n');
                    setFieldValue('masterNodesToAdd', nodeIpArrayByLine);
                  }}
                />
              </FormControl>
              <FormControl
                label="Worker nodes"
                tooltip="Worker nodes execute tasks assigned by the control plane nodes and handle the execution of containers and workloads to keep your applications running smoothly."
                inputId="workerNodesToAdd"
                errors={errors.workerNodesToAdd}
                // reduce the bottom gap so that the test connection button is closer to the input (but still below the front end validation errors)
                className="!mb-0 [&>div>.help-block>p]:!mb-0 [&>div>.help-block]:!mb-0 [&>label]:!pl-0"
              >
                <Field
                  name="workerNodesToAdd"
                  as={NodeAddressInput}
                  type="text"
                  data-cy="microk8sEditForm-workerNodesInput"
                  id="controlPlaneNodes"
                  nodeIPValues={values.workerNodesToAdd}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => {
                    const nodeIpArrayByLine = e.target.value.split('\n');
                    setFieldValue('workerNodesToAdd', nodeIpArrayByLine);
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
                      !!errors.masterNodesToAdd ||
                      !!errors.workerNodesToAdd ||
                      credentialID === undefined
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

              <FormSection title="Actions">
                <div className="flex w-full flex-wrap gap-2">
                  <LoadingButton
                    disabled={!isValid || disableAddNodes}
                    isLoading={isSubmitting || isTestingConnectionOnSubmit}
                    type="button"
                    color="primary"
                    size="small"
                    className="!ml-0"
                    onClick={async () => onAddNodes()}
                    loadingText="Adding nodes"
                  >
                    Add nodes
                  </LoadingButton>
                  <Button
                    type="button"
                    color="default"
                    size="small"
                    onClick={() => {
                      setFieldValue('masterNodesToAdd', ['']);
                      setFieldValue('workerNodesToAdd', ['']);
                      // setTimeout is needed to make sure the validation is run after the values are reset
                      setTimeout(() => {
                        validateField('masterNodesToAdd');
                        validateField('workerNodesToAdd');
                      });
                    }}
                  >
                    Cancel
                  </Button>
                </div>
              </FormSection>
            </Form>
          );

          async function onAddNodes() {
            // if already tested and successful, submit form
            if (isCurrentValuesTested && isSSHTestSuccessful) {
              sendAnalytics();
              submitForm();
              router.stateService.go('kubernetes.cluster');
              return;
            }
            // otherwise, test connection and submit form if the test is successful
            try {
              setIsTestingConnectionOnSubmit(true);
              const [isTestConnectionSuccessful] = await handleTestConnection();
              if (isTestConnectionSuccessful) {
                submitForm();
                sendAnalytics();
                router.stateService.go('kubernetes.cluster');
              }
            } finally {
              setIsTestingConnectionOnSubmit(false);
            }
          }

          function sendAnalytics() {
            const currentMasterNodeCount = nodes?.filter(
              (node) => getRole(node) === 'Control plane'
            ).length;
            const currentWorkerNodeCount = nodes?.filter(
              (node) => getRole(node) === 'Worker'
            ).length;
            trackEvent('scale-up-k8s-cluster', {
              category: 'kubernetes',
              metadata: {
                provider: K8sDistributionType.MICROK8S,
                currentMasterNodeCount,
                currentWorkerNodeCount,
                masterNodesToAddCount: values.masterNodesToAdd.length,
                workerNodesToAddCount: values.workerNodesToAdd.length,
              },
            });
          }

          // handleTestConnection tests the SSH connection to the nodes and returns a boolean (indicating whether the test was successful)
          // and a number (the number of successful tests)
          function handleTestConnection(): Promise<[boolean, number]> {
            return new Promise((resolve) => {
              if (credentialID === undefined) {
                resolve([false, 0]);
                return;
              }
              const combinedNodeIPs = formatNodeIPs([
                ...values.masterNodesToAdd,
                ...values.workerNodesToAdd,
              ]);
              testSSHConnectionMutation.mutate(
                {
                  nodeIPs: combinedNodeIPs,
                  credentialID,
                },
                {
                  onSuccess: (addressResults) => {
                    const [failedAddressResults, successfulAddressResults] =
                      partition(addressResults, (result) => result.error);
                    const isTestSuccessful = failedAddressResults.length === 0;
                    // update the component state with the results of the test
                    setAddressResults(addressResults);
                    setTestedAddressList(combinedNodeIPs);
                    setIsSSHTestSuccessful(isTestSuccessful);
                    setFailedAddressResults(failedAddressResults);
                    // resolve with the results of the test, and the number of successful addresses
                    resolve([
                      isTestSuccessful,
                      successfulAddressResults.length,
                    ]);
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
        }}
      </Formik>
    </WidgetBody>
  );
}
