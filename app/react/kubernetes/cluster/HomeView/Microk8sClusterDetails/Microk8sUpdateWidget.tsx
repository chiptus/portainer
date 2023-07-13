import Kube from '@/assets/ico/kube.svg?c';
import { EnvironmentStatus } from '@/react/portainer/environments/types';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { BetaAlert } from '@/react/portainer/environments/update-schedules/common/BetaAlert';

import { Widget, WidgetTitle } from '@@/Widget';

import { Addons } from '../../microk8s/addons/Addons';

import { ErrorStatus } from './ErrorStatus';
import { UpgradeCluster } from './UpgradeCluster';
import { UpgradeStatus } from './UpgradeStatus';

export function Microk8sUpdateWidget() {
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId);

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle icon={Kube} title="MicroK8s cluster management" />
          <div className="flex flex-col gap-y-5 p-5">
            {environment?.Status === EnvironmentStatus.Error ? (
              <ErrorStatus />
            ) : (
              <>
                <BetaAlert message="Beta feature - so far, MicroK8s cluster management functionality has only been tested in a limited set of scenarios." />
                <UpgradeStatus />
                <UpgradeCluster />
                <Addons />
              </>
            )}
          </div>
        </Widget>
      </div>
    </div>
  );
}
