import { LogInterface } from '@@/LogViewer/types';
import { TIMESTAMP_LENGTH } from '@@/LogViewer/helpers/consts';
import { startsWithTimestamp } from '@@/LogViewer/helpers/commons';
import { prettifyJSONLog } from '@@/LogViewer/helpers/prettifyLogs/prettifyJSONLog';
import { prettifyZeroLog } from '@@/LogViewer/helpers/prettifyLogs/prettifyZeroLog';

function prettifyLog(
  newLogs: LogInterface[],
  log: LogInterface,
  withTimestamp: boolean
) {
  let { line } = log;
  let timestamp = '';

  if (withTimestamp) {
    timestamp = line.substring(0, TIMESTAMP_LENGTH);
    line = line.substring(TIMESTAMP_LENGTH);
  }

  if (!prettifyJSONLog(newLogs, line, timestamp)) {
    if (!prettifyZeroLog(newLogs, line, timestamp)) {
      newLogs.push(log);
    }
  }
}

export function prettifyLogs(logs: LogInterface[]) {
  const newLogs: LogInterface[] = [];

  if (logs.length) {
    const withTimestamp = startsWithTimestamp(logs[0].line);

    for (let i = 0; i < logs.length; i += 1) {
      prettifyLog(newLogs, logs[i], withTimestamp);
    }
  }

  return newLogs;
}
