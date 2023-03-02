import { useFormikContext } from 'formik';
import { Plus, RefreshCw } from 'lucide-react';

import { LoadingButton } from '@@/buttons/LoadingButton';
import { FormSection } from '@@/form-components/FormSection';

import { FormValues } from '../types';

interface Props {
  isSubmitting: boolean;
  onReloadClick: () => void;
  isReloading: boolean;
}

export function ActionsSection({
  isSubmitting,
  onReloadClick,
  isReloading,
}: Props) {
  const { isValid } = useFormikContext<FormValues>();

  return (
    <FormSection title="Actions">
      <div className="mb-3 flex w-full flex-wrap gap-2">
        <LoadingButton
          disabled={!isValid}
          isLoading={isSubmitting}
          loadingText="Provision in progress..."
          icon={Plus}
          className="!ml-0"
        >
          Provision environment
        </LoadingButton>

        <LoadingButton
          type="button"
          color="default"
          onClick={onReloadClick}
          isLoading={isReloading}
          loadingText="Reloading details..."
          icon={RefreshCw}
          className="!ml-0"
        >
          Reload cluster details
        </LoadingButton>
      </div>
    </FormSection>
  );
}
