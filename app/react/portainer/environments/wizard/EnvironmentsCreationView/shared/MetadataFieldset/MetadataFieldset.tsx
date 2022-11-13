import { useField } from 'formik';
import { PropsWithChildren } from 'react';

import { useUser } from '@/react/hooks/useUser';

import { TagSelector } from '@@/TagSelector';
import { FormSection } from '@@/form-components/FormSection';

import { GroupField } from './GroupsField';

export function MetadataFieldset({ children }: PropsWithChildren<unknown>) {
  const [tagProps, , tagHelpers] = useField('meta.tagIds');

  const { isAdmin } = useUser();

  return (
    <FormSection title="Metadata">
      {children}

      <GroupField />

      <TagSelector
        value={tagProps.value}
        allowCreate={isAdmin}
        onChange={(value) => tagHelpers.setValue(value)}
      />
    </FormSection>
  );
}
