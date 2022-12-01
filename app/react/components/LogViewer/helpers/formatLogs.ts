import { formatANSILogs } from '@@/LogViewer/helpers/formatANSILogs';
import {
  LogInterface,
  RawLogsSection,
  RawLogsInterface,
} from '@@/LogViewer/types';
import { prefixLogsSection } from '@@/LogViewer/helpers/prefixLogsSection';
import { postProcessLogs } from '@@/LogViewer/helpers/postProcessLogs';
import { prettifyLogs } from '@@/LogViewer/helpers/prettifyLogs/pretifyLogs';

function stripLogsHeaders(logs: string) {
  // header := [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
  // STREAM_TYPE can be:
  // 0: stdin (is written on stdout)
  // 1: stdout
  // 2: stderr
  // eslint-disable-next-line no-control-regex
  const headerPattern = /[\u0000\u0001\u0002]\u0000{3}.{4}/gs;
  return logs.replace(headerPattern, '').replace(/\r\n/g, '\n');
}

export function formatLogs(logs: RawLogsInterface) {
  const rawLogsSections: RawLogsSection[] =
    typeof logs === 'string' ? [{ logs }] : logs;

  const logSections: LogInterface[][] = [];

  for (let i = 0; i < rawLogsSections.length; i += 1) {
    const rawLogsSection = rawLogsSections[i];

    const stripedLogs = stripLogsHeaders(rawLogsSection.logs);
    let formattedLogs = formatANSILogs(stripedLogs);
    formattedLogs = prettifyLogs(formattedLogs);
    formattedLogs = prefixLogsSection(formattedLogs, rawLogsSection);
    logSections.push(formattedLogs);
  }

  return postProcessLogs(logSections.flat());
}
