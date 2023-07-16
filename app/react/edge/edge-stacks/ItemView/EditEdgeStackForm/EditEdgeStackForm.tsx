import { useCurrentStateAndParams } from '@uirouter/react';

import { useEdgeStack } from '../../queries/useEdgeStack';

import { GitForm } from './GitForm';
import { TextForm } from './TextForm';

export function EditEdgeStackForm() {
  const {
    params: { stackId },
  } = useCurrentStateAndParams();
  const edgeStackQuery = useEdgeStack(stackId);

  if (!edgeStackQuery.data) {
    return null;
  }

  const edgeStack = edgeStackQuery.data;

  if (edgeStack.GitConfig) {
    return <GitForm stack={edgeStack} />;
  }

  return <TextForm edgeStack={edgeStack} />;
}
