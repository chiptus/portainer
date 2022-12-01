import { useContext } from 'react';
import { saveAs } from 'file-saver';
import { Download } from 'lucide-react';

import { Icon } from '@@/Icon';
import { Button } from '@@/buttons';
import {
  LogViewerContext,
  LogViewerContextInterface,
} from '@@/LogViewer/types';
import { NEW_LINE_BREAKER } from '@@/LogViewer/helpers/consts';

export function DownloadLogsButton() {
  const { logs, resourceName } = useContext(
    LogViewerContext
  ) as LogViewerContextInterface;

  function onClick() {
    const data = logs.logs.map((log) => log.line + NEW_LINE_BREAKER);
    const filename = `${resourceName}_logs.txt`;
    saveAs(new Blob(data), filename);
  }

  return (
    <Button onClick={onClick} disabled={!logs.logs.length}>
      <Icon icon={Download} />
      <span>Download logs</span>
    </Button>
  );
}
