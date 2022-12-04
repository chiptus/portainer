import {
  LogInterface,
  LogSpanInterface,
  RawLogsSection,
} from '@@/LogViewer/types';

function addLogSectionName(log: LogInterface, rawLogsSection: RawLogsSection) {
  const newLog = { ...log };
  const { sectionName, sectionNameColor } = rawLogsSection;
  const text = `${sectionName} `;
  newLog.line = `${text}${log.line}`;
  const newSpan: LogSpanInterface = {
    text,
    style: { color: sectionNameColor },
  };
  newLog.spans.unshift(newSpan);

  return newLog;
}

export function prefixLogsSection(
  logs: LogInterface[],
  rawLogsSection: RawLogsSection
) {
  if (rawLogsSection.sectionName) {
    const newLogs: LogInterface[] = [];
    for (let i = 0; i < logs.length; i += 1) {
      newLogs.push(addLogSectionName(logs[i], rawLogsSection));
    }
    return newLogs;
  }

  return logs;
}
