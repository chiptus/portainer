import { useState } from 'react';

import { MTLSCertOptions } from '@/react/portainer/settings/EdgeComputeView/EdgeComputeSettings/types';

import { FileUploadField } from '@@/form-components/FileUpload';
import { FormControl } from '@@/form-components/FormControl';
import { Switch } from '@@/form-components/SwitchField/Switch';
import { TextTip } from '@@/Tip/TextTip';

interface Props {
  values: MTLSCertOptions;
  onChange(value: MTLSCertOptions): void;
}

export function MTLSOptions({ onChange, values }: Props) {
  const [valuesCache, setValuesCache] = useState(values);

  function onChangeField(key: string, newValue: unknown) {
    const newValues = {
      ...values,
      [key]: newValue,
    };
    onChange(newValues);
  }

  function onChangeUseSeparateCert(newValue: boolean) {
    if (newValue) {
      onChange({
        ...valuesCache,
        UseSeparateCert: true,
      });
    } else {
      setValuesCache(values);
      onChange({
        UseSeparateCert: false,
      });
    }
  }

  return (
    <>
      <FormControl
        inputId="use_separate_mtls_cert"
        label="Use separate mTLS cert"
        size="small"
        tooltip=""
        className="mb-1"
      >
        <Switch
          id="use_separete_cert"
          name="name_use_separate_cert"
          checked={!!values.UseSeparateCert}
          onChange={() => onChangeUseSeparateCert(!values.UseSeparateCert)}
        />
      </FormControl>

      <TextTip color="blue">
        Use a specific TLS certificate for mTLS communication
      </TextTip>

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
