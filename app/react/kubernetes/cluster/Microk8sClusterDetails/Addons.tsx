import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { EnvironmentStatus } from '@/react/portainer/environments/types';

import { DetailsTable } from '@@/DetailsTable';
import { WidgetBody } from '@@/Widget';
import { Card } from '@@/Card';

import { useAddons } from './Addons.service';

export function Addons() {
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
    <>
      {environment?.Status !== EnvironmentStatus.Error && (
        <WidgetBody loading={addonsQuery.isLoading}>
          <Card>
            {addonsQuery.isError && (
              <DetailsTable.Row label="Addons" colClassName="w-1/3">
                unable to get addons
              </DetailsTable.Row>
            )}
            {addons && (
              <DetailsTable.Row label="Addons" colClassName="w-1/3">
                {addons.length ? addons.join(', ') : 'No addons installed'}
              </DetailsTable.Row>
            )}
          </Card>
        </WidgetBody>
      )}
    </>
  );
}
