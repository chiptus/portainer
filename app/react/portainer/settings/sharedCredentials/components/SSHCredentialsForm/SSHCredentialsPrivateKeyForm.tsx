import { useState } from 'react';
import { Field, useFormikContext } from 'formik';

import { readFileAsText } from '@/portainer/services/fileUploadReact';
import { notifyError } from '@/portainer/services/notifications';

import { FormControl } from '@@/form-components/FormControl';
import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { FileUploadField } from '@@/form-components/FileUpload';
import { Button } from '@@/buttons';
import { Input } from '@@/form-components/Input';
import { SwitchField } from '@@/form-components/SwitchField';
import { TextArea } from '@@/form-components/Input/Textarea';

import { SSHCredentialFormValues } from '../../types';

import { SSHKeyGenerationModal } from './SSHKeyGenerationModal';

type Props = {
  sshPrivateKeyValue: string;
  passphraseValue: string;
  hasPassphrase?: boolean;
  hasSSHKey?: boolean;
  privateKeyErrors?: string;
  passphraseErrors?: string;
  isEditing?: boolean;
};

export default function SSHCredentialsPrivateKeyForm({
  sshPrivateKeyValue,
  passphraseValue,
  privateKeyErrors,
  passphraseErrors,
  isEditing,
  hasSSHKey,
  hasPassphrase,
}: Props) {
  const [showSSHKeyFields, setShowSSHKeyFields] = useState(
    hasSSHKey || !isEditing
  );
  const [privateKeyFile, setPrivateKeyFile] = useState<File>();
  const [isPassphraseModalOpen, setIsPassphraseModalOpen] = useState(false);
  const { setFieldValue } = useFormikContext<SSHCredentialFormValues>();
  return (
    <div className="flex flex-col">
      <SwitchField
        label="Use SSH key authentication"
        dataCy="cloudSettings-useSshKey"
        labelClass="col-sm-4 col-lg-3"
        fieldClass="!mb-3"
        tooltip="SSH keys are a pair of public and private keys used to authenticate access."
        checked={showSSHKeyFields}
        disabled={isEditing && hasSSHKey}
        onChange={() => setShowSSHKeyFields(!showSSHKeyFields)}
      />
      {showSSHKeyFields && (
        <>
          <FormSectionTitle>SSH Key</FormSectionTitle>
          <div className="mb-4 flex gap-2">
            <Button
              color="default"
              className="!ml-0"
              onClick={() => {
                setIsPassphraseModalOpen(true);
              }}
            >
              Generate new RSA SSH key pair
            </Button>
            <FileUploadField
              inputId="sshKey-privateKeyUpload"
              data-cy="sshKey-privateKeyUpload"
              title="Upload SSH private key"
              className="btn-default"
              value={privateKeyFile}
              onChange={async (file) => {
                try {
                  const maxFileSize = 1024 * 1024; // 1MB
                  const fileText = await readFileAsText(file, maxFileSize);
                  setFieldValue('credentials.privateKey', fileText);
                  setPrivateKeyFile(file);
                } catch (error) {
                  notifyError(
                    'Unable to read private SSH key file as text',
                    error as Error
                  );
                }
              }}
            />
          </div>
          <FormControl
            inputId="passphrase"
            label="SSH private key passphrase"
            size="medium"
            tooltip="If your private key is encrypted, enter the passphrase to decrypt it here."
            errors={passphraseErrors}
          >
            <Field
              as={Input}
              name="credentials.passphrase"
              type="password"
              autoComplete="off"
              id="ssh_passphrase"
              value={passphraseValue}
              placeholder={hasPassphrase ? '*******' : ''}
              data-cy="cloudSettings-SshPassphrase"
            />
          </FormControl>
          <FormControl
            inputId="ssh_private_key"
            label="SSH private key"
            size="medium"
            errors={privateKeyErrors}
          >
            <Field
              as={StyledTextArea}
              name="credentials.privateKey"
              autoComplete="off"
              type="password"
              id="ssh_key"
              value={sshPrivateKeyValue}
              placeholder={
                hasSSHKey
                  ? '*******'
                  : `e.g. \n-----BEGIN RSA PRIVATE KEY----- \nb3BlbnNzaC1... \n-----END RSA PRIVATE KEY-----`
              }
              data-cy="cloudSettings-privateKeyTextArea"
            />
          </FormControl>
          {isPassphraseModalOpen && (
            <SSHKeyGenerationModal
              onDismiss={(
                cont: boolean,
                passphrase?: string,
                generatedPrivateKey?: string
              ) => {
                setIsPassphraseModalOpen(false);
                if (!cont) {
                  return;
                }
                setFieldValue('credentials.passphrase', passphrase);
                setFieldValue('credentials.privateKey', generatedPrivateKey);
              }}
            />
          )}
        </>
      )}
    </div>
  );
}

export function StyledTextArea({ ...props }) {
  return (
    // eslint-disable-next-line react/jsx-props-no-spreading
    <TextArea {...props} className="form-control min-h-[150px] resize-y" />
  );
}
