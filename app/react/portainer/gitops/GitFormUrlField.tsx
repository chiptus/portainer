import { ChangeEvent, useEffect, useLayoutEffect, useRef } from 'react';
import { RefreshCcw } from 'react-feather';
import { useQueryClient } from 'react-query';

import { useCheckRepo } from '@/react/portainer/gitops/queries/useCheckRepo';
import { useDebounce } from '@/portainer/hooks/useDebounce';

import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { TextTip } from '@@/Tip/TextTip';
import { Button } from '@@/buttons';

import { GitFormModel } from './types';

interface Props {
  value: string;
  onChange(value: string): void;
  onChangeRepositoryValid(value: boolean): void;
  onRefreshGitopsCache(): void;
  model: GitFormModel;
}

export function GitFormUrlField({
  value,
  onChange,
  onChangeRepositoryValid,
  onRefreshGitopsCache,
  model,
}: Props) {
  const queryClient = useQueryClient();

  const handleChangeRef = useRef(onChangeRepositoryValid);
  useLayoutEffect(() => {
    handleChangeRef.current = onChangeRepositoryValid;
  });

  const repository = useDebounce(value);
  // eslint-disable-next-line no-nested-ternary
  const creds = model.RepositoryPassword
    ? {
        username: model.RepositoryUsername,
        password: model.RepositoryPassword,
      }
    : model.SelectedGitCredential
    ? { gitCredentialId: model.SelectedGitCredential.id }
    : {};
  const payload = {
    repository,
    ...creds,
  };
  const enabled = !!(repository && repository.length > 0);
  const repoStatusQuery = useCheckRepo(payload, enabled);

  useEffect(() => {
    if (!repoStatusQuery.isLoading && enabled)
      handleChangeRef.current(!repoStatusQuery.isError);
  }, [repoStatusQuery.isError, repoStatusQuery.isLoading, enabled]);

  return (
    <div className="form-group">
      <span className="col-sm-12">
        <TextTip color="blue">You can use the URL of a git repository.</TextTip>
      </span>
      <div className="col-sm-12">
        <FormControl
          label="Repository URL"
          inputId="stack_repository_url"
          errors={repoStatusQuery.error?.message}
          required
        >
          <span className="flex">
            <Input
              value={value}
              type="text"
              name="repoUrlField"
              className="form-control"
              placeholder="https://github.com/portainer/portainer-compose"
              data-cy="component-gitUrlInput"
              required
              onChange={handleChange}
            />

            <Button
              onClick={onRefresh}
              size="medium"
              className="vertical-center"
              color="light"
              icon={RefreshCcw}
              title="refreshGitRepo"
              disabled={!model.RepositoryURLValid}
            />
          </span>
        </FormControl>
      </div>
    </div>
  );

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    onChange(e.target.value);
  }

  function onRefresh() {
    onRefreshGitopsCache();
    queryClient.invalidateQueries(['git_repo_refs', 'git_repo_search_results']);
  }
}
