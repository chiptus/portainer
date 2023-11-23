import { useState } from 'react';

import { useCurrentUser } from '@/react/hooks/useUser';

import { UpdateConfirmDialog } from './UpdateConfirmDialog';
import { LoadingDialog } from './LoadingDialog';
import { NonAdminUpdateDialog } from './NonAdminUpdateDialog';

type Step = 'confirm' | 'loading';

export function UpdateDialog({ onDismiss }: { onDismiss: () => void }) {
  const { isPureAdmin } = useCurrentUser();
  const [currentStep, setCurrentStep] = useState<Step>('confirm');
  const component = getDialog();

  return component;

  function getDialog() {
    if (!isPureAdmin) {
      return <NonAdminUpdateDialog onDismiss={onDismiss} />;
    }

    switch (currentStep) {
      case 'confirm':
        return (
          <UpdateConfirmDialog
            goToLoading={() => setCurrentStep('loading')}
            onDismiss={onDismiss}
          />
        );
      case 'loading':
        return <LoadingDialog />;
      default:
        throw new Error('step type not found');
    }
  }
}
