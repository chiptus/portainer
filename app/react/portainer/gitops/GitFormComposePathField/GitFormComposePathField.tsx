import {
  Combobox,
  ComboboxInput,
  ComboboxList,
  ComboboxOption,
  ComboboxPopover,
} from '@reach/combobox';
import '@reach/combobox/styles.css';
import { ChangeEvent, useState } from 'react';
import clsx from 'clsx';

import { useDebounce } from '@/portainer/hooks/useDebounce';
import { useSearch } from '@/react/portainer/gitops/queries/useSearch';

import { FormControl } from '@@/form-components/FormControl';
import { TextTip } from '@@/Tip/TextTip';

import { GitFormModel } from '../types';

import styles from './GitFormComposePathField.module.css';

interface Props {
  value: string;
  onChange(value: string): void;
  isCompose: boolean;
  model: GitFormModel;
  isDockerStandalone: boolean;
}

export function GitFormComposePathField({
  value,
  onChange,
  isCompose,
  model,
  isDockerStandalone,
}: Props) {
  const [searchTerm, setSearchTerm] = useState('');
  const debouncedSearchValue = useDebounce(searchTerm);

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
    repository: model.RepositoryURL,
    keyword: debouncedSearchValue,
    reference: model.RepositoryReferenceName,
    ...creds,
  };
  const enabled = Boolean(
    model.RepositoryURL && model.RepositoryURLValid && debouncedSearchValue
  );
  const { data: searchResult } = useSearch(payload, enabled);

  return (
    <div className={clsx('form-group', styles.root)}>
      <span className="col-sm-12">
        <TextTip color="blue">
          <span>
            Indicate the path to the {isCompose ? 'Compose' : 'Manifest'} file
            from the root of your repository (requires a yaml, yml, json, or hcl
            file extension).
          </span>
          {isDockerStandalone && (
            <span>
              {' '}
              To enable rebuilding of an image if already present on Docker
              standalone environments, include
              <code>pull_policy: build</code> in your compose file as per{' '}
              <a href="https://docs.docker.com/compose/compose-file/#pull_policy">
                Docker documentation
              </a>
              .
            </span>
          )}
        </TextTip>
      </span>
      <div className="col-sm-12">
        <FormControl
          label={isCompose ? 'Compose path' : 'Manifest path'}
          inputId="stack_repository_path"
          required
        >
          <Combobox
            aria-label="compose"
            onSelect={onSelect}
            data-cy="component-gitComposeInput"
          >
            <ComboboxInput
              className="form-control"
              onChange={handleChange}
              placeholder={isCompose ? 'docker-compose.yml' : 'manifest.yml'}
              value={value}
            />
            {searchResult && searchResult.length > 0 && searchTerm !== '' && (
              <ComboboxPopover>
                <ComboboxList>
                  {searchResult.map((result: string, index: number) => (
                    <ComboboxOption key={index} value={result} />
                  ))}
                </ComboboxList>
              </ComboboxPopover>
            )}
          </Combobox>
        </FormControl>
      </div>
    </div>
  );

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    setSearchTerm(e.target.value);
    onChange(e.target.value);
  }

  function onSelect(value: string) {
    setSearchTerm('');
    onChange(value);
  }
}
