import { useField } from 'formik';

import { EdgeGroupsSelector } from '@/react/edge/edge-stacks/components/EdgeGroupsSelector';

import { FormValues } from './types';

export function EdgeGroupsField() {
  const [{ value }, { error }, { setValue }] =
    useField<FormValues['groupIds']>('groupIds');

  return (
    <EdgeGroupsSelector
      value={value}
      onChange={(value) => setValue(value)}
      error={error}
      horizontal
    />
  );
}
