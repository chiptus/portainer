import Kube from '@/assets/ico/kube.svg?c';
import { EnvironmentStatus } from '@/react/portainer/environments/types';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';

import { Widget, WidgetTitle } from '@@/Widget';

import { useAddonsQuery } from '../../microk8s/addons.service';

import { Addons } from './Addons';
import { ErrorStatus } from './ErrorStatus';
import { UpgradeCluster } from './UpgradeCluster';
import { UpgradeStatus } from './UpgradeStatus';

export function Microk8sUpdateWidget() {
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId);

  const statusQuery = useAddonsQuery(environment?.Id, environment?.Status);
  const { currentVersion, kubernetesVersions } = statusQuery.data || {};

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle icon={Kube} title="MicroK8s cluster details" />
          <div className="flex flex-col gap-y-5 p-5">
            {environment?.Status === EnvironmentStatus.Error ? (
              <ErrorStatus />
            ) : (
              <>
                <UpgradeStatus />
                <UpgradeCluster
                  currentVersion={currentVersion}
                  kubernetesVersions={kubernetesVersions}
                  statusQuery={{
                    isLoading: statusQuery.isLoading,
                    isError: statusQuery.isError,
                  }}
                />
                {currentVersion && <Addons currentVersion={currentVersion} />}
              </>
            )}
          </div>
        </Widget>
      </div>
    </div>
  );
}
