import { Formik } from 'formik';

import { EdgeStack } from '../../../types';

import { FormValues } from './types';
import { formValidation } from './validation';
import { InnerForm } from './InnerForm';

interface Props {
  edgeStack: EdgeStack;
  isSubmitting: boolean;
  onSubmit: (values: FormValues) => void;
  onEditorChange: (content: string) => void;
  fileContent: string;
  allowKubeToSelectCompose: boolean;
}

export function TextForm({
  isSubmitting,
  edgeStack,
  onSubmit,
  onEditorChange,
  fileContent,
  allowKubeToSelectCompose,
}: Props) {
  const formValues: FormValues = {
    edgeGroups: edgeStack.EdgeGroups,
    deploymentType: edgeStack.DeploymentType,
    privateRegistryId: edgeStack.Registries?.[0],
    content: fileContent,
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
      onSubmit={onSubmit}
      validationSchema={formValidation()}
    >
      <InnerForm
        edgeStack={edgeStack}
        isSubmitting={isSubmitting}
        onEditorChange={onEditorChange}
        allowKubeToSelectCompose={allowKubeToSelectCompose}
        versionOptions={versionOptions}
      />
    </Formik>
  );
}
