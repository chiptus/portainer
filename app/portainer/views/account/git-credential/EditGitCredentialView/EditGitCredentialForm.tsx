import { useCurrentStateAndParams, useRouter } from '@uirouter/react';

import { useUser } from '@/portainer/hooks/useUser';

import {
  GitCredential,
  GitCredentialFormValues,
  UpdateGitCredentialPayload,
} from '../types';
import {
  useGitCredentials,
  useUpdateGitCredentialMutation,
} from '../gitCredential.service';
import { GitCredentialForm } from '../components/GitCredentialForm';

type Props = {
  gitCredential: GitCredential;
};

export function EditGitCredentialForm({ gitCredential }: Props) {
  const router = useRouter();
  const { params } = useCurrentStateAndParams();
  const currentUser = useUser();
  const gitCredentialsQuery = useGitCredentials(currentUser.user.Id);
  const gitCredentialNames = gitCredentialsQuery.data || [];

  const updateGitCredentialMutation = useUpdateGitCredentialMutation();

  const defaultInitialValues = {
    name: gitCredential.name,
    username: gitCredential.username,
    password: '',
  };

  return (
    <GitCredentialForm
      isEditing
      isLoading={updateGitCredentialMutation.isLoading}
      onSubmit={onSubmit}
      gitCredentialNames={gitCredentialNames
        .filter((x) => x.id !== gitCredential.id)
        .map((x) => x.name)}
      initialValues={defaultInitialValues}
    />
  );

  function onSubmit(values: GitCredentialFormValues) {
    const newCredentials: UpdateGitCredentialPayload = {
      ...values,
    };

    updateGitCredentialMutation.mutate({
      credential: newCredentials,
      id: params.id,
      userId: currentUser.user.Id,
    });

    router.stateService.go('portainer.account');
  }
}
