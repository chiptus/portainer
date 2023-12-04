import { useRouter } from '@uirouter/react';

import { useUser } from '@/react/hooks/useUser';

import { GitCredentialFormValues } from '../types';
import { useGitCredentials } from '../git-credentials.service';
import { GitCredentialForm } from '../components/GitCredentialForm';
import { useCreateGitCredentialMutation } from '../queries/useCreateGitCredentialsMutation';

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
    createGitCredentialMutation.mutate(
      {
        ...values,
        userId: currentUser.user.Id,
      },
      {
        onSuccess: () => {
          if (routeOnSuccess) {
            router.stateService.go(routeOnSuccess);
          }
        },
      }
    );
  }
}
