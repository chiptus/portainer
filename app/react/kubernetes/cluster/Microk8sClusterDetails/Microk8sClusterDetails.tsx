import Kube from '@/assets/ico/kube.svg?c';
import { EnvironmentStatus } from '@/react/portainer/environments/types';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';

import { Widget, WidgetTitle } from '@@/Widget';

import { Addons } from './Addons';
import { ErrorStatus } from './ErrorStatus';

export function Microk8sClusterDetails() {
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId);

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle icon={Kube} title="MicroK8s cluster details" />
          {environment?.Status === EnvironmentStatus.Error ? (
            <ErrorStatus />
          ) : (
            <Addons />
          )}
        </Widget>
      </div>
    </div>
  );
}
