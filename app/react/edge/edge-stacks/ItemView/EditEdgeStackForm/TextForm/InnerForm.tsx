import { useFormikContext, Form } from 'formik';
import { useState, useEffect } from 'react';

import { baseEdgeStackWebhookUrl } from '@/portainer/helpers/webhookHelper';
import { EnvironmentType } from '@/react/portainer/environments/types';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { EdgeGroupsSelector } from '@/react/edge/edge-stacks/components/EdgeGroupsSelector';
import { EdgeStackDeploymentTypeSelector } from '@/react/edge/edge-stacks/components/EdgeStackDeploymentTypeSelector';
import { getEdgeStackFile } from '@/react/edge/edge-stacks/queries/useEdgeStackFile';
import { EdgeStack, DeploymentType } from '@/react/edge/edge-stacks/types';
import { WebhookSettings } from '@/react/portainer/gitops/AutoUpdateFieldset/WebhookSettings';
import { useConfirmBeforeExit } from '@/react/hooks/useConfirmBeforeExitEditor';

import { TextTip } from '@@/Tip/TextTip';
import { LoadingButton } from '@@/buttons';
import { EnvironmentVariablesPanel } from '@@/form-components/EnvironmentVariablesFieldset';
import { FormError } from '@@/form-components/FormError';
import { FormSection } from '@@/form-components/FormSection';
import { SwitchField } from '@@/form-components/SwitchField';

import { atLeastTwo } from '../atLeastTwo';
import { useValidateEnvironmentTypes } from '../useValidateEnvironmentTypes';

import { PrivateRegistryFieldsetWrapper } from './PrivateRegistryFieldsetWrapper';
import { ComposeForm } from './ComposeForm';
import { NomadForm } from './NomadForm';
import { KubernetesForm } from './KubernetesForm';
import { FormValues } from './types';
import { useCachedContent } from './useCachedValue';
import { useKubeAllowedToCompose } from './useKubeAllowedToCompose';

const forms = {
  [DeploymentType.Compose]: ComposeForm,
  [DeploymentType.Kubernetes]: KubernetesForm,
  [DeploymentType.Nomad]: NomadForm,
};

export function InnerForm({
  edgeStack,
  isSubmitting,
  versionOptions,
  originalContent,
  skipEditorExitCheck,
}: {
  edgeStack: EdgeStack;
  isSubmitting: boolean;
  versionOptions: number[];
  originalContent: string;
  skipEditorExitCheck: boolean;
}) {
  const { values, setFieldValue, isValid, errors, setFieldError, dirty } =
    useFormikContext<FormValues>();
  const { getCachedContent, setContentCache } = useCachedContent();
  const { hasType } = useValidateEnvironmentTypes(values.edgeGroups);
  const allowKubeToSelectCompose = useKubeAllowedToCompose();
  const [selectedVersion, setSelectedVersion] = useState(versionOptions[0]);
  useConfirmBeforeExit(originalContent, values.content, skipEditorExitCheck);

  useEffect(() => {
    if (selectedVersion !== versionOptions[0]) {
      setFieldValue('rollbackTo', selectedVersion);
    } else {
      setFieldValue('rollbackTo', undefined);
    }
  }, [selectedVersion, setFieldValue, versionOptions]);

  const hasKubeEndpoint = hasType(EnvironmentType.EdgeAgentOnKubernetes);
  const hasDockerEndpoint = hasType(EnvironmentType.EdgeAgentOnDocker);
  const hasNomadEndpoint = hasType(EnvironmentType.EdgeAgentOnNomad);

  const DeploymentForm = forms[values.deploymentType];

  return (
    <Form className="form-horizontal">
      <EdgeGroupsSelector
        value={values.edgeGroups}
        onChange={(value) => setFieldValue('edgeGroups', value)}
        error={errors.edgeGroups}
      />

      {atLeastTwo(hasKubeEndpoint, hasDockerEndpoint, hasNomadEndpoint) && (
        <TextTip>
          There are no available deployment types when there is more than one
          type of environment in your edge group selection (e.g. Kubernetes and
          Docker environments). Please select edge groups that have environments
          of the same type.
        </TextTip>
      )}

      {values.deploymentType === DeploymentType.Compose && hasKubeEndpoint && (
        <FormError>
          Edge groups with kubernetes environments no longer support compose
          deployment types in Portainer. Please select edge groups that only
          have docker environments when using compose deployment types.
        </FormError>
      )}

      <EdgeStackDeploymentTypeSelector
        allowKubeToSelectCompose={allowKubeToSelectCompose}
        value={values.deploymentType}
        hasDockerEndpoint={hasType(EnvironmentType.EdgeAgentOnDocker)}
        hasKubeEndpoint={hasType(EnvironmentType.EdgeAgentOnKubernetes)}
        hasNomadEndpoint={hasType(EnvironmentType.EdgeAgentOnNomad)}
        onChange={(value) => {
          setFieldValue('content', getCachedContent(value));
          setFieldValue('deploymentType', value);
        }}
      />

      <DeploymentForm
        hasKubeEndpoint={hasType(EnvironmentType.EdgeAgentOnKubernetes)}
        handleContentChange={handleContentChange}
        versionOptions={versionOptions}
        handleVersionChange={handleVersionChange}
      />

      {isBE && (
        <>
          <FormSection title="Webhooks">
            <div className="form-group">
              <div className="col-sm-12">
                <SwitchField
                  label="Create an Edge stack webhook"
                  checked={values.webhookEnabled}
                  labelClass="col-sm-3 col-lg-2"
                  onChange={(value) => setFieldValue('webhookEnabled', value)}
                  tooltip="Create a webhook (or callback URI) to automate the update of this stack. Sending a POST request to this callback URI (without requiring any authentication) will pull the most up-to-date version of the associated image and re-deploy this stack."
                />
              </div>
            </div>

            {edgeStack.Webhook && (
              <>
                <WebhookSettings
                  baseUrl={baseEdgeStackWebhookUrl()}
                  value={edgeStack.Webhook}
                  docsLink=""
                />

                <TextTip color="orange">
                  Sending environment variables to the webhook is updating the
                  stack with the new values. New variables names will be added
                  to the stack and existing variables will be updated.
                </TextTip>
              </>
            )}
          </FormSection>

          <PrivateRegistryFieldsetWrapper
            value={values.privateRegistryId}
            onChange={(value) => setFieldValue('privateRegistryId', value)}
            isValid={isValid}
            values={values}
            stackName={edgeStack.Name}
            onFieldError={(error) => setFieldError('privateRegistryId', error)}
            error={errors.privateRegistryId}
          />

          {values.deploymentType === DeploymentType.Compose && (
            <>
              <EnvironmentVariablesPanel
                onChange={(value) => setFieldValue('envVars', value)}
                values={values.envVars}
                errors={errors.envVars}
              />

              <div className="form-group">
                <div className="col-sm-12">
                  <SwitchField
                    checked={values.prePullImage}
                    name="prePullImage"
                    label="Pre-pull images"
                    tooltip="When enabled, redeployment will be executed when image(s) is pulled successfully"
                    labelClass="col-sm-3 col-lg-2"
                    onChange={(value) => setFieldValue('prePullImage', value)}
                  />
                </div>
              </div>

              <div className="form-group">
                <div className="col-sm-12">
                  <SwitchField
                    checked={values.retryDeploy}
                    name="retryDeploy"
                    label="Retry deployment"
                    tooltip="When enabled, this will allow edge agent keep retrying deployment if failure occur"
                    labelClass="col-sm-3 col-lg-2"
                    onChange={(value) => setFieldValue('retryDeploy', value)}
                  />
                </div>
              </div>
            </>
          )}
        </>
      )}

      <FormSection title="Actions">
        <div className="form-group">
          <div className="col-sm-12">
            <LoadingButton
              className="!ml-0"
              size="small"
              disabled={!isValid || !dirty}
              isLoading={isSubmitting}
              loadingText="Update in progress..."
            >
              Update the stack
            </LoadingButton>
          </div>
        </div>
      </FormSection>
    </Form>
  );

  function handleContentChange(type: DeploymentType, content: string) {
    setFieldValue('content', content);
    setContentCache(type, content);
  }

  async function handleVersionChange(newVersion: number) {
    const fileContent = await getEdgeStackFile(edgeStack.Id, newVersion).catch(
      () => ''
    );
    if (fileContent) {
      if (versionOptions.length > 1) {
        if (newVersion < versionOptions[0]) {
          setSelectedVersion(newVersion);
        } else {
          setSelectedVersion(versionOptions[0]);
        }
      }
      handleContentChange(values.deploymentType, fileContent);
    }
  }
}
