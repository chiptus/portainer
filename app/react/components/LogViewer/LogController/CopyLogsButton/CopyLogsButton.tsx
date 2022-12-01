import { Check, Copy } from 'lucide-react';
import { useCallback, useContext } from 'react';

import { Icon } from '@@/Icon';
import { Button } from '@@/buttons';
import {
  LogViewerContext,
  LogViewerContextInterface,
} from '@@/LogViewer/types';
import { useCopy } from '@@/buttons/CopyButton/useCopy';
import { NEW_LINE_BREAKER } from '@@/LogViewer/helpers/consts';

export function CopyLogsButton() {
  const { logs } = useContext(LogViewerContext) as LogViewerContextInterface;

  const copyText = useCallback(
    () => logs.logs.map((log) => log.line + NEW_LINE_BREAKER).join(''),
    [logs.logs]
  );

  const { handleCopy, copiedSuccessfully } = useCopy(copyText, 1000);
  const disabled = !logs.logs.length || copiedSuccessfully;

  return (
    <Button onClick={handleCopy} disabled={disabled} color="default">
      <Icon icon={Copy} />
      <span>{copiedSuccessfully ? 'Copied' : 'Copy'}</span>
      {copiedSuccessfully && <Icon icon={Check} mode="success" />}
    </Button>
  );
}
