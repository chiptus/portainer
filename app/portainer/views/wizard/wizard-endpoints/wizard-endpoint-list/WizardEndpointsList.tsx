import clsx from 'clsx';

import { Widget, WidgetBody, WidgetTitle } from '@/portainer/components/widget';
import {
  environmentTypeIcon,
  endpointTypeName,
  stripProtocol,
} from '@/portainer/filters/filters';
import { Environment } from '@/portainer/environments/types';
import { EdgeIndicator } from '@/portainer/home/EnvironmentList/EnvironmentItem';
import { isEdgeEnvironment } from '@/portainer/environments/utils';

import styles from './wizard-endpoint-list.module.css';

interface Props {
  environments: Environment[];
}

export function WizardEndpointsList({ environments }: Props) {
  return (
    <Widget>
      <WidgetTitle icon="fa-plug" title="Connected Environments" />
      <WidgetBody>
        {environments.map((environment) => (
          <div className={styles.wizardListWrapper} key={environment.Id}>
            <div className={styles.wizardListImage}>
              <i
                aria-hidden="true"
                className={clsx(
                  'space-right',
                  environmentTypeIcon(environment.Type)
                )}
              />
            </div>
            <div className={styles.wizardListTitle}>{environment.Name}</div>
            <div className={styles.wizardListSubtitle}>
              URL: {stripProtocol(environment.URL)}
            </div>
            <div className={styles.wizardListType}>
              Type: {endpointTypeName(environment.Type)}
            </div>
            {isEdgeEnvironment(environment.Type) && (
              <div className={styles.wizardListEdgeStatus}>
                <EdgeIndicator
                  edgeId={environment.EdgeID}
                  checkInInterval={environment.EdgeCheckinInterval}
                  queryDate={environment.QueryDate}
                  lastCheckInDate={environment.LastCheckInDate}
                />
              </div>
            )}
          </div>
        ))}
      </WidgetBody>
    </Widget>
  );
}
