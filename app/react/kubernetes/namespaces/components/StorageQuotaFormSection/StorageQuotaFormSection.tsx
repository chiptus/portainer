import { FormikErrors } from 'formik';

import { StorageClass } from '@/react/portainer/environments/types';

import { FormSection } from '@@/form-components/FormSection';
import { TextTip } from '@@/Tip/TextTip';

import { StorageQuotaItem } from './StorageQuotaItem';
import { StorageQuotaFormValues } from './types';

interface Props {
  values: StorageQuotaFormValues[];
  onChange: (value: StorageQuotaFormValues[]) => void;
  storageClasses: StorageClass[];
  errors?:
    | string
    | string[]
    | FormikErrors<StorageQuotaFormValues>[]
    | undefined;
}

export function StorageQuotaFormSection({
  values,
  onChange,
  storageClasses,
  errors,
}: Props) {
  return (
    <FormSection title="Storage">
      <TextTip color="blue">
        Quotas can be set on each storage option to prevent users from exceeding
        a specific threshold when deploying applications. You can set a quota to
        0 to effectively prevent the usage of a specific storage option inside
        this namespace.
      </TextTip>

      {storageClasses.map((storageClass, index) => (
        <StorageQuotaItem
          key={storageClass.Name}
          value={values[index]}
          onChange={(value) => {
            const newStorageQuotas = [...values];
            newStorageQuotas[index] = value;
            onChange(newStorageQuotas);
          }}
          storageClass={storageClass}
          errors={errors?.[index]}
        />
      ))}
    </FormSection>
  );
}
