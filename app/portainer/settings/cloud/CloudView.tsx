import { PageHeader } from '@/portainer/components/PageHeader';
import { react2angular } from '@/react-tools/react2angular';
import { Widget, WidgetBody, WidgetTitle } from '@/portainer/components/widget';

import { CloudSettingsForm } from './CloudSettingsForm';

export function CloudView() {
  return (
    <>
      <PageHeader
        title="Cloud settings"
        breadcrumbs={[
          { label: 'Settings', link: 'portainer.settings' },
          { label: 'Cloud' },
        ]}
      />

      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetTitle title="API keys" icon="fa-cloud" />
            <WidgetBody>
              <CloudSettingsForm reroute showDigitalOcean showLinode showCivo />
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}

export const CloudViewAngular = react2angular(CloudView, []);
