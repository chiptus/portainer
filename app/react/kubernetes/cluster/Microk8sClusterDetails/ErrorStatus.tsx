import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';

import { TextTip } from '@@/Tip/TextTip';
import { WidgetBody } from '@@/Widget';

export function ErrorStatus() {
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId);

  if (!environment?.StatusMessage) {
    return null;
  }

  return (
    <WidgetBody>
      <TextTip childrenWrapperClassName="ml-2 text-muted" color="red">
        {environment?.StatusMessage.Summary && (
          <div className="font-medium">{environment.StatusMessage.Summary}</div>
        )}
        {environment?.StatusMessage.Detail}
      </TextTip>
    </WidgetBody>
  );
}
