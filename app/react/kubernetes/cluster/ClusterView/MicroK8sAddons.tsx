import Kube from '@/assets/ico/kube.svg?c';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { EnvironmentStatus } from '@/react/portainer/environments/types';

import { Widget, WidgetTitle, WidgetBody } from '@@/Widget';
import { DetailsTable } from '@@/DetailsTable';
import { TextTip } from '@@/Tip/TextTip';

import { useAddons } from './microk8sAddons.service';

export function Microk8sAddons() {
  const environmentId = useEnvironmentId();
  const environmentQuery = useEnvironment(environmentId);
  const { data: environment } = environmentQuery;
  const addonsQuery = useAddons(
    environment?.Id,
    environment?.CloudProvider.CredentialID,
    environment?.Status
  );
  const addons = addonsQuery.data?.addons;

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle icon={Kube} title="MicroK8s cluster details" />
          {environment?.Status === EnvironmentStatus.Error && (
            <WidgetBody>
              <TextTip childrenWrapperClassName="ml-2 text-muted" color="red">
                {environment.StatusMessage.Summary && (
                  <div className="font-medium">
                    {environment.StatusMessage.Summary}
                  </div>
                )}
                {environment.StatusMessage.Detail}
              </TextTip>
            </WidgetBody>
          )}
          {environment?.Status !== EnvironmentStatus.Error && (
            <WidgetBody loading={addonsQuery.isLoading}>
              <DetailsTable>
                {addonsQuery.isError && (
                  <DetailsTable.Row label="Addons" colClassName="w-1/2">
                    unable to get addons
                  </DetailsTable.Row>
                )}
                {addons && (
                  <DetailsTable.Row label="Addons" colClassName="w-1/2">
                    {addons.length ? addons.join(', ') : 'No addons installed'}
                  </DetailsTable.Row>
                )}
              </DetailsTable>
            </WidgetBody>
          )}
        </Widget>
      </div>
    </div>
  );
}
