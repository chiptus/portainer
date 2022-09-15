import { useRouter } from '@uirouter/react';

import { useUser } from '@/portainer/hooks/useUser';

import { CreateGitCredentialPayload, GitCredentialFormValues } from '../types';
import {
  useCreateGitCredentialMutation,
  useGitCredentials,
} from '../gitCredential.service';
import { GitCredentialForm } from '../components/GitCredentialForm';

type Props = {
  routeOnSuccess?: string;
};

export function CreateGitCredentialForm({ routeOnSuccess }: Props) {
  const router = useRouter();
  const currentUser = useUser();

  const createGitCredentialMutation = useCreateGitCredentialMutation();
  const gitCredentialsQuery = useGitCredentials(currentUser.user.Id);
  const gitCredentialNames = gitCredentialsQuery.data || [];

  return (
    <GitCredentialForm
      isLoading={createGitCredentialMutation.isLoading}
      onSubmit={onSubmit}
      gitCredentialNames={gitCredentialNames.map((x) => x.name)}
    />
  );

  function onSubmit(values: GitCredentialFormValues) {
    const payload: CreateGitCredentialPayload = {
      ...values,
      userId: currentUser.user.Id,
    };
    createGitCredentialMutation.mutate(payload, {
      onSuccess: () => {
        if (routeOnSuccess) {
          router.stateService.go(routeOnSuccess);
        }
      },
    });
  }
}
