import Kube from '@/assets/ico/kube.svg?c';

import { Widget, WidgetTitle } from '@@/Widget';

import { Addons } from './Addons';
import { ErrorStatus } from './ErrorStatus';

export function Microk8sClusterDetails() {
  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle icon={Kube} title="MicroK8s cluster details" />
          <ErrorStatus />
          <Addons />
        </Widget>
      </div>
    </div>
  );
}
