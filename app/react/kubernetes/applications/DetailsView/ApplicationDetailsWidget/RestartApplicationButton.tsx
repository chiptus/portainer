import { RefreshCw } from 'lucide-react';
import { useRouter } from '@uirouter/react';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { notifySuccess, notifyError } from '@/portainer/services/notifications';
import { Authorized } from '@/react/hooks/useUser';

import { confirm } from '@@/modals/confirm';
import { ModalType } from '@@/modals';
import { buildConfirmButton } from '@@/modals/utils';
import { Button } from '@@/buttons';
import { Icon } from '@@/Icon';

import { useRolloutRestartApplicationMutation } from '../../application.queries';
import { AppKind } from '../../types';

type Props = {
  environmentId: EnvironmentId;
  namespace: string;
  appName: string;
  appKind?: AppKind;
};

export function RestartApplicationButton({
  environmentId,
  namespace,
  appName,
  appKind,
}: Props) {
  const router = useRouter();
  const rolloutRestartMutation = useRolloutRestartApplicationMutation(
    environmentId,
    namespace,
    appName
  );

  return (
    <Authorized authorizations="K8sPodRestart">
      <Button
        type="button"
        size="small"
        color="light"
        className="!ml-0"
        disabled={rolloutRestartMutation.isLoading}
        onClick={() => {
          restartApplication();
        }}
        data-cy="k8sAppDetail-restartButton"
      >
        <Icon icon={RefreshCw} className="mr-1" />
        Rolling restart
      </Button>
    </Authorized>
  );

  async function restartApplication() {
    const confirmed = await confirm({
      title: 'Are you sure?',
      modalType: ModalType.Warn,
      confirmButton: buildConfirmButton('Rolling restart'),
      message:
        'A rolling restart of the application will be performed, with pods replaced one by one, which should avoid downtime. Do you wish to continue?',
    });
    if (!confirmed || !appKind) {
      return;
    }
    rolloutRestartMutation.mutateAsync(
      { kind: appKind },
      {
        onSuccess: () => {
          notifySuccess('Success', 'Application successfully restarted');
          router.stateService.reload();
        },
        onError: (error) => {
          notifyError(
            'Failure',
            error as Error,
            'Unable to restart the application'
          );
        },
      }
    );
  }
}
