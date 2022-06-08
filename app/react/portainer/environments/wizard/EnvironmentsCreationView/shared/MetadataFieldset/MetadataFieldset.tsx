import { useField } from 'formik';
import { PropsWithChildren } from 'react';

import { TagSelector } from '@/react/components/TagSelector';
import { useUser } from '@/portainer/hooks/useUser';
import { FormSection } from '@/portainer/components/form-components/FormSection';

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
