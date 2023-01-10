import { useMutation } from 'react-query';
import { useEffect } from 'react';
import { Laptop } from 'lucide-react';

import { generateKey } from '@/react/portainer/environments/environment.service/edge';
import { EdgeScriptForm } from '@/react/edge/components/EdgeScriptForm';
import { commandsTabs } from '@/react/edge/components/EdgeScriptForm/scripts';

import { Widget, WidgetBody, WidgetTitle } from '@@/Widget';
import { TextTip } from '@@/Tip/TextTip';

import { useSettings } from '../../queries';

import { AutoEnvCreationSettingsForm } from './AutoEnvCreationSettingsForm';

const commands = {
  linux: [
    commandsTabs.k8sLinux,
    commandsTabs.swarmLinux,
    commandsTabs.standaloneLinux,
    commandsTabs.nomadLinux,
  ],
  win: [commandsTabs.swarmWindows, commandsTabs.standaloneWindow],
};

export function AutomaticEdgeEnvCreation() {
  const edgeKeyMutation = useGenerateKeyMutation();
  const { mutate: generateKey } = edgeKeyMutation;
  const settingsQuery = useSettings();

  const url = settingsQuery.data?.EdgePortainerUrl;

  const settings = settingsQuery.data;
  const edgeKey = edgeKeyMutation.data;
  const edgeComputeConfigurationOK = !!(
    settings &&
    settings.EnableEdgeComputeFeatures &&
    settings.EdgePortainerUrl &&
    settings.Edge.TunnelServerAddress
  );

  useEffect(() => {
    if (edgeComputeConfigurationOK) {
      generateKey();
    }
  }, [generateKey, edgeComputeConfigurationOK]);

  if (!settingsQuery.data) {
    return null;
  }

  return (
    <Widget>
      <WidgetTitle icon={Laptop} title="Automatic Edge Environment Creation" />
      <WidgetBody>
        {!edgeComputeConfigurationOK ? (
          <TextTip color="orange">
            In order to use this feature, please make sure that Edge Compute
            features are turned on and that you have properly configured the
            Portainer API server URL and tunnel server address.
          </TextTip>
        ) : (
          <>
            <AutoEnvCreationSettingsForm settings={settings} />

            {edgeKeyMutation.isLoading ? (
              <div>Generating key for {url} ... </div>
            ) : (
              edgeKey && (
                <>
                  <hr />
                  <EdgeScriptForm
                    edgeInfo={{ key: edgeKey }}
                    commands={commands}
                    isNomadTokenVisible
                  />
                </>
              )
            )}
          </>
        )}
      </WidgetBody>
    </Widget>
  );
}

// using mutation because we want this action to run only when required
function useGenerateKeyMutation() {
  return useMutation(generateKey);
}
