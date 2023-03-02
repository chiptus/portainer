import { MTLSCertOptions } from '@/react/portainer/settings/EdgeComputeView/DeploymentSyncOptions/types';

import { FileUploadField } from '@@/form-components/FileUpload';
import { FormControl } from '@@/form-components/FormControl';
import { Switch } from '@@/form-components/SwitchField/Switch';
import { TextTip } from '@@/Tip/TextTip';

interface Props {
  values: MTLSCertOptions;
  onChange(value: MTLSCertOptions): void;
}

export function MTLSOptions({ onChange, values }: Props) {
  function onChangeField(key: string, newValue: unknown) {
    const newValues = {
      ...values,
      [key]: newValue,
    };
    onChange(newValues);
  }

  return (
    <>
      <TextTip color="blue">
        Use a specific TLS certificate for mTLS communication
      </TextTip>

      <FormControl
        inputId="use_separate_mtls_cert"
        label="Use separate mTLS cert"
        size="small"
        tooltip=""
      >
        <Switch
          id="use_separete_cert"
          name="name_use_separate_cert"
          checked={!!values.UseSeparateCert}
          onChange={() =>
            onChangeField('UseSeparateCert', !values.UseSeparateCert)
          }
        />
      </FormControl>

      {values.UseSeparateCert && (
        <>
          <FormControl label="TLS CA certificate" inputId="ca-cert-field">
            <FileUploadField
              inputId="ca-cert-field"
              onChange={(file) => onChangeField('CaCertFile', file)}
              value={values.CaCertFile}
            />
          </FormControl>
          <FormControl label="TLS certificate" inputId="cert-field">
            <FileUploadField
              inputId="cert-field"
              onChange={(file) => onChangeField('CertFile', file)}
              value={values.CertFile}
            />
          </FormControl>
          <FormControl label="TLS key" inputId="tls-key-field">
            <FileUploadField
              inputId="tls-key-field"
              onChange={(file) => onChangeField('KeyFile', file)}
              value={values.KeyFile}
            />
          </FormControl>
        </>
      )}
    </>
  );
}
