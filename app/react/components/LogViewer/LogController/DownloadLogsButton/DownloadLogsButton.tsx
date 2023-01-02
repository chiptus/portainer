import { saveAs } from 'file-saver';
import { Download } from 'lucide-react';

import { Icon } from '@@/Icon';
import { Button } from '@@/buttons';
import { NEW_LINE_BREAKER } from '@@/LogViewer/helpers/consts';
import { useLogViewerContext } from '@@/LogViewer/context';

export function DownloadLogsButton() {
  const { logs, resourceName } = useLogViewerContext();

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
