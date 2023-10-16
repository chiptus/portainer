import { FormikErrors } from 'formik';
import { Database } from 'lucide-react';

import { StorageClass } from '@/react/portainer/environments/types';
import { FeatureId } from '@/react/portainer/feature-flags/enums';

// import { Select } from '@@/form-components/Input/Select';
import { Icon } from '@@/Icon';
import { FormControl } from '@@/form-components/FormControl';
import { FormError } from '@@/form-components/FormError';
import { FormSectionTitle } from '@@/form-components/FormSectionTitle';
import { InputGroup } from '@@/form-components/InputGroup';
import { SwitchField } from '@@/form-components/SwitchField';
import { Input } from '@@/form-components/Input';
import { isErrorType } from '@@/form-components/formikUtils';
import { Select } from '@@/form-components/ReactSelect';

import { StorageQuotaFormValues } from './types';

const sizeUnitOptions: {
  label: string;
  value: StorageQuotaFormValues['sizeUnit'];
}[] = [
  { label: 'MB', value: 'M' },
  { label: 'GB', value: 'G' },
  { label: 'TB', value: 'T' },
];

type Props = {
  value: StorageQuotaFormValues;
  onChange: (value: StorageQuotaFormValues) => void;
  storageClass: StorageClass;
  errors?: string | FormikErrors<StorageQuotaFormValues>;
};

export function StorageQuotaItem({
  value: storageQuota,
  onChange,
  storageClass,
  errors,
}: Props) {
  const storageQuotaErrors = isErrorType<StorageQuotaFormValues>(errors)
    ? errors
    : null;

  return (
    <div key={storageClass.Name}>
      <FormSectionTitle>
        <div className="vertical-center text-muted inline-flex gap-1 align-top">
          <Icon icon={Database} className="!mt-0.5 flex-none" />
          <span>{storageClass.Name}</span>
        </div>
      </FormSectionTitle>
      <hr className="mt-2 mb-0 w-full" />
      <div className="form-group">
        <div className="col-sm-12">
          <SwitchField
            data-cy="k8sNamespaceEdit-storageClassQuota"
            disabled={false}
            label="Enable quota"
            labelClass="col-sm-3 col-lg-2"
            fieldClass="pt-2"
            checked={storageQuota.enabled || false}
            onChange={(enabled) => onChange({ ...storageQuota, enabled })}
            featureId={FeatureId.K8S_RESOURCE_POOL_STORAGE_QUOTA}
          />
        </div>
      </div>
      {storageQuota.enabled && (
        <FormControl label="Maximum usage" required>
          <InputGroup className="col-sm-3 min-w-fit flex">
            <Input
              required
              type="number"
              min={0}
              onChange={(event) =>
                onChange({
                  ...storageQuota,
                  size: event.target.value,
                })
              }
              value={storageQuota.size ?? ''}
              data-cy="namespaceCreate-maxUsage"
              placeholder="e.g. 20"
              className="col-sm-3"
            />
            <Select
              options={sizeUnitOptions}
              defaultValue={sizeUnitOptions[0]}
              value={sizeUnitOptions.find(
                (option) => option.value === storageQuota.sizeUnit
              )}
              onChange={(option) =>
                onChange({
                  ...storageQuota,
                  sizeUnit: option?.value ?? 'M',
                })
              }
              className="!w-20 [&>div]:!w-20"
              data-cy="namespaceCreate-storageQuotaUnitSelect"
            />
          </InputGroup>
          {storageQuotaErrors?.size && (
            <FormError>{storageQuotaErrors.size}</FormError>
          )}
        </FormControl>
      )}
    </div>
  );
}
