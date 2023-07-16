import { Formik } from 'formik';
// import { useState } from 'react';
import { useRouter } from '@uirouter/react';
import _ from 'lodash';

import { DeploymentType, EdgeStack } from '@/react/edge/edge-stacks/types';
import { createWebhookId } from '@/portainer/helpers/webhookHelper';
import { confirmStackUpdate } from '@/react/common/stacks/common/confirm-stack-update';
import { notifySuccess } from '@/portainer/services/notifications';
import { useEdgeStackFile } from '@/react/edge/edge-stacks/queries/useEdgeStackFile';

import { FormValues } from './types';
import { useUpdateEdgeStackMutation } from './useUpdateEdgeStackMutation';
import { InnerForm } from './InnerForm';
import { validation } from './validation';

interface Props {
  edgeStack: EdgeStack;
}

export function TextForm({ edgeStack }: Props) {
  const router = useRouter();
  const fileQuery = useEdgeStackFile(edgeStack.Id);
  const deployMutation = useUpdateEdgeStackMutation();
  // const [skipConfirmExitCheck, setConfirmExitCheck] = useState(false);
  if (typeof fileQuery.data !== 'string') {
    return null;
  }

  const formValues: FormValues = {
    edgeGroups: edgeStack.EdgeGroups,
    deploymentType: edgeStack.DeploymentType,
    privateRegistryId: edgeStack.Registries?.[0],
    content: fileQuery.data,
    useManifestNamespaces: edgeStack.UseManifestNamespaces,
    prePullImage: edgeStack.PrePullImage,
    retryDeploy: edgeStack.RetryDeploy,
    webhookEnabled: !!edgeStack.Webhook,
    envVars: edgeStack.EnvVars || [],
    rollbackTo: undefined,
  };

  const versionOptions: number[] = edgeStack.StackFileVersion
    ? [edgeStack.StackFileVersion]
    : [];
  if (edgeStack.PreviousDeploymentInfo?.FileVersion > 0) {
    versionOptions.push(edgeStack.PreviousDeploymentInfo?.FileVersion);
  }

  return (
    <Formik
      initialValues={formValues}
      onSubmit={handleSubmit}
      validationSchema={validation}
    >
      <InnerForm
        edgeStack={edgeStack}
        isSubmitting={deployMutation.isLoading}
        // originalContent={fileQuery.data}
        versionOptions={versionOptions}
        // skipEditorExitCheck={skipConfirmExitCheck}
      />
    </Formik>
  );

  async function handleSubmit(values: FormValues) {
    let rePullImage = false;
    if (values.deploymentType === DeploymentType.Compose) {
      const defaultToggle = values.prePullImage;
      const result = await confirmStackUpdate(
        'Do you want to force an update of the stack?',
        defaultToggle
      );
      if (!result) {
        return;
      }

      rePullImage = result.pullImage;
    }

    const updateVersion = !!(
      fileQuery.data !== values.content ||
      values.privateRegistryId !== edgeStack.Registries[0] ||
      values.useManifestNamespaces !== edgeStack.UseManifestNamespaces ||
      values.prePullImage !== edgeStack.PrePullImage ||
      values.retryDeploy !== edgeStack.RetryDeploy ||
      _.differenceWith(values.envVars, edgeStack.EnvVars || [], _.isEqual)
        .length > 0 ||
      rePullImage
    );

    deployMutation.mutate(
      {
        id: edgeStack.Id,
        stackFileContent: values.content,
        edgeGroups: values.edgeGroups,
        deploymentType: values.deploymentType,
        registries: values.privateRegistryId ? [values.privateRegistryId] : [],
        useManifestNamespaces: values.useManifestNamespaces,
        prePullImage: values.prePullImage,
        updateVersion,
        rePullImage,
        retryDeploy: values.retryDeploy,
        webhook: values.webhookEnabled
          ? edgeStack.Webhook || createWebhookId()
          : '',
        envVars: values.envVars,
        rollbackTo: values.rollbackTo ? values.rollbackTo : undefined,
      },
      {
        onSuccess() {
          // setConfirmExitCheck(true);
          notifySuccess('Success', 'Stack successfully deployed');
          router.stateService.go('edge.stacks');
        },
      }
    );
  }
}
