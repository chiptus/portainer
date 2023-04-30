import { useCurrentUser } from '@/react/hooks/useUser';
import { useExperimentalSettings } from '@/react/portainer/settings/queries';

export function useCanDisplayChatbot(): boolean {
  const { data: settings } = useExperimentalSettings();
  const { user } = useCurrentUser();

  return !!(
    settings &&
    settings.experimentalFeatures.OpenAIIntegration &&
    user.OpenAIApiKey
  );
}
