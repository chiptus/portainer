import {LogInterface} from "@@/LogViewer/types";


export function postProcessLogs(logs: LogInterface[]) {
  for (let i = 0; i < logs.length; i += 1) {
    const log = logs[i];
    log.lineNumber = i + 1;
    log.lowerLine = log.line.toLowerCase();
  }

  return logs;
}
