import { ChangeEvent, useEffect, useLayoutEffect, useRef } from 'react';

import { useGitRefs } from '@/react/portainer/gitops/queries/useGitRefs';

import { FormControl } from '@@/form-components/FormControl';
import { Select } from '@@/form-components/Input';
import { TextTip } from '@@/Tip/TextTip';

import { GitFormModel } from './types';

interface Props {
  value: string;
  onChange(value: string): void;
  model: GitFormModel;
}

export function GitFormRefField({ value, onChange, model }: Props) {
  const handleChangeRef = useRef(onChange);
  useLayoutEffect(() => {
    handleChangeRef.current = onChange;
  });

  let creds = {};
  if (model.RepositoryAuthentication) {
    if (model.RepositoryPassword) {
      creds = {
        username: model.RepositoryUsername,
        password: model.RepositoryPassword,
      };
    } else if (model.SelectedGitCredential) {
      creds = { gitCredentialID: model.SelectedGitCredential.id };
    }
  }
  const payload = {
    repository: model.RepositoryURL,
    ...creds,
    stackID: model.StackID,
  };

  const enabled = Boolean(model.RepositoryURL && model.RepositoryURLValid);
  const { data: refs } = useGitRefs(payload, enabled, (refs) => {
    let options = [{ value: 'refs/heads/main', label: 'refs/heads/main' }];

    if (refs?.length > 0) {
      // put refs/heads/main first if it is present in repository
      if (refs.indexOf('refs/heads/main') > 0) {
        refs.splice(refs.indexOf('refs/heads/main'), 1);
        refs.unshift('refs/heads/main');
      }

      options = refs.map((t: string) => ({
        value: t,
        label: t,
      }));
    }

    return options;
  });

  useEffect(() => {
    if (refs && !refs?.some((ref) => ref.value === value)) {
      handleChangeRef.current(refs[0].value);
    }
  }, [value, refs]);

  return (
    <div className="form-group">
      <span className="col-sm-12">
        <TextTip color="blue">
          Specify a reference of the repository using the following syntax:
          branches with <code>refs/heads/branch_name</code> or tags with{' '}
          <code>refs/tags/tag_name</code>.
        </TextTip>
      </span>
      <div className="col-sm-12">
        <FormControl
          label="Repository reference"
          inputId="stack_repository_reference_name"
          required
        >
          <Select
            value={value}
            options={
              refs || [{ value: 'refs/heads/main', label: 'refs/heads/main' }]
            }
            onChange={handleChange}
            data-cy="component-gitRefInput"
          />
        </FormControl>
      </div>
    </div>
  );

  function handleChange(e: ChangeEvent<HTMLSelectElement>) {
    onChange(e.target.value);
  }
}
