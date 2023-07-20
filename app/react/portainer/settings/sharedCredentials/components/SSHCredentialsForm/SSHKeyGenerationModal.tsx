import { ChangeEvent, useState } from 'react';
import { Download } from 'lucide-react';

import { notifyError } from '@/portainer/services/notifications';

import { Button, CopyButton, LoadingButton } from '@@/buttons';
import { Modal } from '@@/modals/Modal';
import { FormControl } from '@@/form-components/FormControl';
import { Input } from '@@/form-components/Input';
import { TextArea } from '@@/form-components/Input/Textarea';
import { DownloadButton } from '@@/buttons/DownloadButton';
import { FormError } from '@@/form-components/FormError';

import { useGenerateSSHKeyMutation } from './generateSSHKey.service';

export function SSHKeyGenerationModal({
  onDismiss,
}: {
  onDismiss: (
    cont: boolean,
    passphrase?: string,
    generatedPrivateKey?: string
  ) => void;
}) {
  const [passphrase, setPassphrase] = useState('');
  const [generatedPrivateKey, setGeneratedPrivateKey] = useState('');
  const [generatedPublicKey, setGeneratedPublicKey] = useState('');
  const generateSSHKeyMutation = useGenerateSSHKeyMutation();

  return (
    <Modal aria-label="Generate SSH key pair" size="lg">
      <Modal.Header title="Generate RSA SSH key pair" />
      <Modal.Body>
        <FormControl
          label="Passphrase (optional)"
          setTooltipHtmlMessage
          tooltip={
            <div>
              <p>
                Encrypting your private key with a passphrase adds an extra
                layer of security to the generated SSH private key.
              </p>
              <p>
                Note down the passphrase now if you want to use the private SSH
                key outside of Portainer.
              </p>
            </div>
          }
          className="flex items-center [&>div]:!pr-0 [&>label]:mb-0 [&>label]:pl-0"
        >
          <Input
            name="generate-ssh-passphrase"
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              setPassphrase(e.target.value)
            }
            type="password"
            maxLength={100}
            autoComplete="off"
            value={passphrase}
            data-cy="generateSSHModal-passphraseInput"
          />
        </FormControl>
        <div className="flex w-full items-center">
          <LoadingButton
            isLoading={generateSSHKeyMutation.isLoading}
            loadingText="Generating SSH key pair"
            size="medium"
            color="secondary"
            onClick={() => {
              generateSSHKeyMutation.mutate(passphrase, {
                onSuccess: (data) => {
                  setGeneratedPrivateKey(data.private);
                  setGeneratedPublicKey(data.public);
                },
                onError: (error) => {
                  notifyError(
                    'Unable to generate SSH key pair',
                    error as Error
                  );
                },
              });
            }}
          >
            Generate SSH key pair
          </LoadingButton>
        </div>
        <div className="my-4 w-full">
          {generatedPrivateKey && (
            <FormControl label="Private key" className="[&>label]:pl-0">
              <TextArea
                value={generatedPrivateKey}
                className="min-h-[96px] resize-y"
                disabled
                data-cy="generateSSHModal-privateKeyTextArea"
              />
              <CopyDownloadActions
                fileName="private_key"
                fileContent={generatedPrivateKey}
                keyType="private"
              />
            </FormControl>
          )}
          {generatedPublicKey && (
            <div className="w-full">
              <FormControl label="Public key" className="[&>label]:pl-0">
                <>
                  <TextArea
                    value={generatedPublicKey}
                    className="min-h-[64px] resize-y"
                    disabled
                    data-cy="generateSSHModal-publicKeyTextArea"
                  />
                  <CopyDownloadActions
                    fileName="public_key.pub"
                    fileContent={generatedPublicKey}
                    keyType="public"
                  />
                </>
              </FormControl>
              <FormError className="!items-start [&>svg]:mt-0.5">
                Please save your public and private keys now. In addition, add
                the generated public key to the authorized_keys file on each
                node you wish to provision for a cluster.
              </FormError>
            </div>
          )}
        </div>
      </Modal.Body>
      <Modal.Footer>
        <div className="flex w-full justify-between gap-2">
          <Button
            color="default"
            size="medium"
            onClick={() => onDismiss(false)}
          >
            Cancel
          </Button>
          <Button
            color="primary"
            size="medium"
            disabled={!generatedPrivateKey}
            onClick={() => onDismiss(true, passphrase, generatedPrivateKey)}
          >
            Continue
          </Button>
        </div>
      </Modal.Footer>
    </Modal>
  );
}

interface CopyDownloadActionsProps {
  fileName: string;
  fileContent: string;
  keyType: 'public' | 'private';
}

function CopyDownloadActions({
  fileName,
  fileContent,
  keyType,
}: CopyDownloadActionsProps) {
  return (
    <div className="mt-2.5 flex w-full gap-2">
      <DownloadButton
        fileContent={fileContent}
        fileName={fileName}
        color="default"
        size="small"
        options={{ type: 'text/text;charset=utf-8' }}
        icon={Download}
        data-cy={`generateSSHModal-download${keyType}KeyButton`}
      >
        Download {keyType} key
      </DownloadButton>
      <CopyButton copyText={fileContent} color="default">
        Copy {keyType} key value
      </CopyButton>
    </div>
  );
}
