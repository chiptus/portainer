import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';

import { TextTip } from '@@/Tip/TextTip';

export function ErrorStatus() {
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId);

  if (!environment?.StatusMessage) {
    return null;
  }

  return (
    <TextTip childrenWrapperClassName="ml-2 text-muted" color="red">
      {environment?.StatusMessage.summary && (
        <div className="font-medium">{environment.StatusMessage.summary}</div>
      )}
      {environment?.StatusMessage.detail}
    </TextTip>
  );
}
