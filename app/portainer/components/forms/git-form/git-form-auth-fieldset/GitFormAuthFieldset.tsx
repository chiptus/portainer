import { ChangeEvent } from 'react';

import { UserGitCredential } from '@/portainer/models/user';

import { SwitchField } from '@@/form-components/SwitchField';
import { Input } from '@@/form-components/Input';
import { FormControl } from '@@/form-components/FormControl';
import { Checkbox } from '@@/form-components/Checkbox';
import { Select } from '@@/form-components/ReactSelect';
import './GitFormAuthFieldset.css';
import { Icon } from '@@/Icon';

interface Props {
  repositoryAuthentication: boolean;
  repositoryUsername: string;
  repositoryPassword: string;
  newCredentialName: string;
  newCredentialNameExist: boolean;
  newCredentialNameInvalid: boolean;
  gitCredentials: UserGitCredential[];
  saveCredential: boolean;
  showAuthExplanation?: boolean;
  selectedGitCredential: UserGitCredential;
  onSelectGitCredential: () => void;
  onChangeRepositoryAuthentication: (checked: boolean) => void;
  onChangeRepositoryUsername: (value: string) => void;
  onChangeRepositoryPassword: (value: string) => void;
  onChangeSaveCredential: (checked: boolean) => void;
  onChangeNewCredentialName: (value: string) => void;
}

export function GitFormAuthFieldset({
  repositoryAuthentication,
  repositoryUsername,
  repositoryPassword,
  newCredentialName,
  newCredentialNameExist,
  newCredentialNameInvalid,
  gitCredentials,
  selectedGitCredential,
  saveCredential,
  showAuthExplanation,
  onSelectGitCredential,
  onChangeRepositoryAuthentication,
  onChangeRepositoryUsername,
  onChangeRepositoryPassword,
  onChangeSaveCredential,
  onChangeNewCredentialName,
}: Props) {
  return (
    <div className="git-form-auth-form">
      <div className="form-group">
        <div className="col-sm-12">
          <SwitchField
            label="Authentication"
            name="authentication"
            checked={repositoryAuthentication}
            onChange={onChangeRepositoryAuthentication}
            data-cy="'component-gitAuthToggle'"
          />
        </div>
      </div>

      {repositoryAuthentication && (
        <>
          {showAuthExplanation && (
            <div
              className="small text-warning"
              style={{ margin: '5px 0 15px 0' }}
            >
              <Icon feather icon="alert-circle" mode="warning" />
              <span className="text-muted space-left">
                Enabling authentication will store the credentials and it is
                advisable to use a git service account
              </span>
            </div>
          )}
          <div className="form-group">
            <div className="col-sm-12">
              <FormControl label="Git Credentials" inputId="users-selector">
                <Select
                  placeholder="select git credential or fill in below"
                  value={gitCredentials.find(
                    (gitCredential) => gitCredential === selectedGitCredential
                  )}
                  options={gitCredentials}
                  getOptionLabel={(gitCredential) => gitCredential.name}
                  getOptionValue={(gitCredential) => gitCredential.id}
                  onChange={onSelectGitCredential}
                  isClearable
                  noOptionsMessage={() => 'no saved credentials'}
                />
              </FormControl>
            </div>
          </div>
          <div className="form-group">
            <div className="col-sm-12">
              <FormControl label="Username">
                <Input
                  value={repositoryUsername}
                  name="repository_username"
                  placeholder={selectedGitCredential ? '' : 'git username'}
                  onChange={(e: ChangeEvent<HTMLInputElement>) =>
                    onChangeRepositoryUsername(e.target.value)
                  }
                  data-cy="component-gitUsernameInput"
                  readOnly={!!selectedGitCredential}
                />
              </FormControl>
            </div>
          </div>
          <div className="form-group personal-access-token">
            <div className="col-sm-12">
              <FormControl
                label="Personal Access Token"
                tooltip="Provide a personal access token or password"
              >
                <Input
                  type="password"
                  value={repositoryPassword}
                  name="repository_password"
                  placeholder="*******"
                  onChange={(e: ChangeEvent<HTMLInputElement>) =>
                    onChangeRepositoryPassword(e.target.value)
                  }
                  data-cy="component-gitPasswordInput"
                  readOnly={!!selectedGitCredential}
                />
              </FormControl>
            </div>
          </div>
          {!selectedGitCredential && repositoryPassword && (
            <div className="form-group save-credential">
              <div className="col-sm-12">
                <FormControl label="">
                  <Checkbox
                    id="repository-save-credential"
                    label="save credential"
                    checked={saveCredential}
                    onChange={(e) => onChangeSaveCredential(e.target.checked)}
                    className="save-credential-check-box"
                  />
                  <Input
                    value={newCredentialName}
                    name="new_credential_name"
                    placeholder="credential name"
                    className="save-credential-name"
                    onChange={(e) => onChangeNewCredentialName(e.target.value)}
                    disabled={!saveCredential}
                  />
                  {newCredentialNameExist && (
                    <div className="small text-danger mb-5">
                      This name is already been used, please try another one
                    </div>
                  )}
                  {newCredentialNameInvalid && (
                    <div className="small text-danger mb-5">
                      This field must consist of lower case alphanumeric
                      characters,&apos;_&apos; or &apos;-&apos; (e.g.
                      &apos;my-name&apos;, or &apos;abc-123&apos;).
                    </div>
                  )}
                  {saveCredential && (
                    <div className="small text-warning">
                      <Icon feather icon="alert-circle" mode="primary" />
                      <span className="text-muted space-left">
                        This git credential can be managed through your account
                        page
                      </span>
                    </div>
                  )}
                </FormControl>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
