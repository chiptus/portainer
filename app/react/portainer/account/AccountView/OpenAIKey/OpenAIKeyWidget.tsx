import { Bot } from 'lucide-react';

import { useCurrentUser } from '@/react/hooks/useUser';
import { useExperimentalSettings } from '@/react/portainer/settings/queries';

import { Widget, WidgetBody, WidgetTitle } from '@@/Widget';

import { OpenAIKeyForm } from './OpenAIKeyForm';

export function OpenAIKeyWidget() {
  const { user } = useCurrentUser();
  const { data: settings } = useExperimentalSettings();

  if (!settings || !settings.experimentalFeatures.OpenAIIntegration) {
    return null;
  }

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetTitle icon={Bot} title="OpenAI integration" />
          <WidgetBody>
            <OpenAIKeyForm user={user} />
          </WidgetBody>
        </Widget>
      </div>
    </div>
  );
}
