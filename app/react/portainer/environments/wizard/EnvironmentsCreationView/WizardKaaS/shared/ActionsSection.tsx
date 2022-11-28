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
      <div className="form-group">
        <div className="col-sm-12">
          <LoadingButton
            disabled={!isValid}
            isLoading={isSubmitting}
            loadingText="Provision in progress..."
            icon={Plus}
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
          >
            Reload cluster details
          </LoadingButton>
        </div>
      </div>
    </FormSection>
  );
}
