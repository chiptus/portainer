import { EdgeStack } from '@/react/edge/edge-stacks/types';

import { FormValues } from './TextForm/types';
import { GitForm } from './GitForm';
import { TextForm } from './TextForm/TextForm';

interface Props {
  edgeStack: EdgeStack;
  isSubmitting: boolean;
  onSubmit: (values: FormValues) => void;
  onEditorChange: (content: string) => void;
  fileContent: string;
  allowKubeToSelectCompose: boolean;
}

export function EditEdgeStackForm({
  isSubmitting,
  edgeStack,
  onSubmit,
  onEditorChange,
  fileContent,
  allowKubeToSelectCompose,
}: Props) {
  if (edgeStack.GitConfig) {
    return <GitForm stack={edgeStack} />;
  }

  return (
    <TextForm
      allowKubeToSelectCompose={allowKubeToSelectCompose}
      edgeStack={edgeStack}
      fileContent={fileContent}
      isSubmitting={isSubmitting}
      onEditorChange={onEditorChange}
      onSubmit={onSubmit}
    />
  );
}
